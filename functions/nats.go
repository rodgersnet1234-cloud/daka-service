package functions

import (
	"daka/model"
	"daka/utils"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

// HandleTask simulates task processing like inserting into DB
func HandleTask(db *sql.DB, task string) bool {
	log.Println("Processing:", task)
	phone := task
	userInfo, err := GetPhone(db, phone)
	if err != nil {
		utils.Log("❌ Failed to find user %d: %v", userInfo.ID, err)
		return false
	}
	//now we get
	package_logs, package_error := RetrievePackageTasksCustom(db, userInfo.ID)
	if package_error != nil {
		utils.Log("❌ Package Error %d: %v", userInfo.ID, err)
		return false
	}
	amount := 0
	for _, pkg := range package_logs {
		amountStr, _ := strconv.Atoi(pkg.Amount)
		amount += amountStr
	}
	// dvAmount := float64(amount / 10)
	// fmt.Println("printed above")
	// UpdateUserAV1(db, userInfo.ID, dvAmount)
	// fmt.Println("added personal av")

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("Failed to load timezone: %v", err)
	}
	today := time.Now().In(loc).Format("2006-01-02")
	used, used_error := GetUsedPackageLogsIDs(db, userInfo.ID, today)
	if used_error != nil {
		UpdateProcessed(db, phone, today)
		//UpdateProcessed(db2, phone, today)
		utils.Log("❌ Retriving used error %d: %v", userInfo.ID, err)
		return false
	}
	if len(used) != 0 {
		//	UpdateProcessed(db, phone, today)
		//UpdateProcessed(db2, phone, today)
		log.Println("❌ Already Processed For Today")
		return false
	}
	fmt.Println(package_logs)
	fmt.Println(used)
	remaining := FilterByExcludingIDs(package_logs, used)
	fmt.Println(remaining)
	fmt.Println("remaining above")
	const maxWorkers = 10

	total_amount := 0.0
	for _, pkg := range remaining {

		amount, err := strconv.ParseFloat(pkg.Amount, 64)
		if err != nil {
			utils.Log(
				"⚠️ PackageLog amount parse error for user %d: %v (value: %s)",
				userInfo.ID, err, pkg.Amount,
			)
			amount = 0
		}

		total_amount += Round(amount/10, 0)

		if err1, str := ProcessPackageLog(db, pkg, userInfo, today); err1 != nil {
			recordFailedPackage(db, pkg, userInfo, str, today, err1)
			log.Println("❌ Error processing package:", err1)
		}
	}

	UpdateProcessed(db, phone, today)
	//	UpdateProcessed(db2, phone, today)
	_, err2 := UpdateUserAV(db, userInfo.ID, total_amount)
	if err2 != nil {
		recordFailedPackage(db, remaining[0], userInfo, "UpdateUserAV", today, err2)
		log.Println("❌ Error processing package:", err2)
	}
	// err2 = UpdateUserAV3(db2, userInfo.ID, avBalance)
	// if err2 != nil {
	// 	recordFailedPackage(db, remaining[0], userInfo, "UpdateUserAV", today, err2)
	// 	log.Println("❌ Error processing package:", err2)
	// }
	checkExists, err2 := Exists(db, userInfo.ID)
	if err2 != nil {
		checkExists = 0
	}
	//UpdateDownGroupCounts(userInfo.ID, db, db2, total_amount)

	if checkExists == 0 {
		fmt.Println("group not exist")
		UpdateDownGroupCounts(userInfo.ID, db, total_amount)
	} else {
		fmt.Println("group exists")
		UpdateGroup(db, userInfo.ID, total_amount)
		//UpdateGroup(db2, userInfo.ID, total_amount)

	}
	utils.Log("🔄 Finished UpdateDownGroupCounts for user %d with amount %.0f", userInfo.ID, total_amount)
	if checkExists == 0 {
		fmt.Println("group1 not exist")
		UpdateParentCounts(db, userInfo.ID)
	} else {
		fmt.Println("group exists2")
		users, err := GetGroup(db, userInfo.ID)
		if err != nil {
			fmt.Println("print error")
			UpdateParentCounts(db, userInfo.ID)
		} else {

			var ug sync.WaitGroup
			//var uu sync.Mutex
			users = append(users, userInfo.ID)
			// const maxWorkers = 10
			jobxs := make(chan int)

			for i := 0; i < maxWorkers; i++ {
				ug.Add(1)
				go func() {
					defer ug.Done()
					for userId := range jobxs {
						user, err := GetUserByID(db, userId)
						if err != nil {
							utils.Log("❌ Failed to find user %d: %v", userId, err)
						} else {
							DailyCalculateUserLevel(db, user)
						}

					}
				}()
			}

			for _, pkg := range users {
				jobxs <- pkg
			}
			close(jobxs)

			ug.Wait()
		}
	}

	return true
}
func UpdateProcessed(db *sql.DB, phone string, date string) error {
	_, err := db.Exec("UPDATE test_ka SET already = 1  WHERE phone = ? and day = ?", phone, date)
	return err
}

func UpdateGroup(db *sql.DB, userId int, amount float64) error {
	_, err := db.Exec("UPDATE users u JOIN user_groups ug ON u.id = ug.ancestor_id SET u.group_av = u.group_av + ? WHERE ug.user_id = ?", amount, userId)
	return err
}

func recordFailedPackage(db *sql.DB, pkg model.PackageTask, user model.RUser, str string, day string, err error) {
	_, e := db.Exec(`
		INSERT INTO failed_packages (package_id, user_id,error, error_message, day)
		VALUES (?, ?, ?, ?, ?)
	`, pkg.ID, user.ID, str, err.Error(), day)
	if e != nil {
		log.Printf("⚠️ Could not record failed package %d: %v", pkg.ID, e)
	}
}

func Exists(db *sql.DB, userId int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM user_groups WHERE user_id = ?"
	err := db.QueryRow(query, userId).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 1, nil
	}
	return 0, nil
}

func GetID(db *sql.DB, phone string) (int, error) {
	var id int
	query := "SELECT id FROM users WHERE phone = ? LIMIT 1"
	err := db.QueryRow(query, phone).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no user found with phone %s", phone)
		}
		return 0, err
	}
	return id, nil
}

func BuRetrieve(db *sql.DB, userID int, day string) ([]model.PackageTask, error) {
	query := "SELECT id, package_id, amount FROM package_logs WHERE user_id = ? AND active = 1 AND add_time != ?"
	rows, err := db.Query(query, userID, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.PackageTask

	for rows.Next() {
		var id int
		var packageID int
		var amount string

		if err := rows.Scan(&id, &packageID, &amount); err != nil {
			return nil, err
		}

		taskType := "ai"
		if packageID <= 3 {
			taskType = "duanju"
		}

		tasks = append(tasks, model.PackageTask{
			Task:   taskType,
			Amount: amount,
			ID:     id,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetGroup(db *sql.DB, userID int) ([]int, error) {
	query := "SELECT ancestor_id FROM user_groups WHERE user_id = ?"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		users = append(users, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func RetrievePackageTasksCustom(db *sql.DB, userID int) ([]model.PackageTask, error) {
	query := "SELECT id, package_id, amount FROM package_logs WHERE user_id = ? AND active = 1 and day != '2025-12-16' and day != '2025-12-15'"
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.PackageTask

	for rows.Next() {
		var id int
		var packageID int
		var amount string

		if err := rows.Scan(&id, &packageID, &amount); err != nil {
			return nil, err
		}

		taskType := "ai"
		if packageID <= 3 {
			taskType = "duanju"
		}

		tasks = append(tasks, model.PackageTask{
			Task:   taskType,
			Amount: amount,
			ID:     id,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func GetTemp(db *sql.DB, userID int, date string) (int, error) {
	query := "SELECT SUM(amount) FROM user_active WHERE user_id = ? AND day = ? AND id > 300000"
	rows, err := db.Query(query, userID, date)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	amount := 0
	var used []int
	for rows.Next() {
		var packageLogID int
		if err := rows.Scan(&amount); err != nil {
			return amount, err
		}
		used = append(used, packageLogID)
	}

	if err = rows.Err(); err != nil {
		return amount, err
	}

	return amount, nil
}

func GetUsedPackageLogsIDs(db *sql.DB, userID int, date string) ([]int, error) {
	query := "SELECT package_logs_id FROM user_active WHERE user_id = ? AND day = ?"
	rows, err := db.Query(query, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var used []int
	for rows.Next() {
		var packageLogID int
		if err := rows.Scan(&packageLogID); err != nil {
			return nil, err
		}
		used = append(used, packageLogID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return used, nil
}
func GetCalcUsers(db *sql.DB, date string) ([]int, error) {
	query := "SELECT id FROM users WHERE id != 1  AND added = 0 ORDER BY id ASC LIMIT 20"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var used []int
	for rows.Next() {
		var packageLogID int
		if err := rows.Scan(&packageLogID); err != nil {
			return nil, err
		}
		used = append(used, packageLogID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return used, nil
}

func GetUserPhones(db *sql.DB, date string) ([]string, error) {
	query := "SELECT phone FROM test_ka WHERE retried = 0 and already = 0 AND day = ? ORDER BY id ASC LIMIT 20"
	rows, err := db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var used []string
	for rows.Next() {
		var packageLogID string
		if err := rows.Scan(&packageLogID); err != nil {
			return nil, err
		}
		used = append(used, packageLogID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return used, nil
}

func FilterByExcludingIDs(data []model.PackageTask, excludedIDs []int) []model.PackageTask {
	// Convert excludedIDs slice into a lookup map for fast O(1) access
	excludedMap := make(map[int]bool)
	for _, id := range excludedIDs {
		excludedMap[id] = true
	}

	var result []model.PackageTask
	for _, item := range data {
		if !excludedMap[item.ID] {
			result = append(result, item)
		}
	}

	return result
}
