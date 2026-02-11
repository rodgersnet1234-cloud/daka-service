package functions

import (
	"daka/config"
	"daka/model"
	"daka/utils"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
)

func UpdateParentCounts(db *sql.DB, currentUserID int) error {
	user, err := GetUserByID(db, currentUserID)
	if err != nil {
		utils.Log("❌ Failed to find user %d: %v", currentUserID, err)
		return err
	}

	// Do your processing for the current user
	DailyCalculateUserLevel(db, user)

	// Check if there's a valid inviter
	if user.Invited.Valid {
		// Recurse to the inviter
		return UpdateParentCounts(db, int(user.Invited.Int64))
	}

	// No inviter found — recursion ends
	return nil
}

func DownlinePlay(db *sql.DB) {
	user, err := GetUserByID(db, 2029)
	if err != nil {
		utils.Log("❌ Failed to find user %d: %v", user.ID, err)
		return
	}
	DailyCalculateUserLevel(db, user)
}

func GetUserByID(db *sql.DB, userID int) (model.RUser, error) {
	query := `
		SELECT 
			id, name, phone, package_value, package_level, packages, level,
			downline_count, invited, av, group_av, shouyi, jifen
		FROM users
		WHERE id = ?
	`

	var u model.RUser
	err := db.QueryRow(query, userID).Scan(
		&u.ID, &u.Name, &u.Phone, &u.PackageValue, &u.PackageLevel,
		&u.Packages, &u.Level, &u.DownlineCount, &u.Invited,
		&u.AV, &u.GroupAV, &u.Shouyi, &u.Jifen,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return u, fmt.Errorf("user with ID %d not found", userID)
		}
		return u, fmt.Errorf("query failed: %w", err)
	}

	return u, nil
}

func GetPhone(db *sql.DB, phone string) (model.RUser, error) {
	query := `
		SELECT 
			id, name, phone, package_value, package_level, packages, level,
			downline_count, invited, av, group_av, jifen, shouyi
		FROM users
		WHERE phone = ?
	`

	var u model.RUser
	err := db.QueryRow(query, phone).Scan(
		&u.ID, &u.Name, &u.Phone, &u.PackageValue, &u.PackageLevel,
		&u.Packages, &u.Level, &u.DownlineCount, &u.Invited,
		&u.AV, &u.GroupAV, &u.Jifen, &u.Shouyi,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return u, fmt.Errorf("user with ID %v not found", phone)
		}
		return u, fmt.Errorf("query failed: %w", err)
	}

	return u, nil
}
func MaxSplitSum(users []int) int {
	max := users[0]

	for _, val := range users[1:] {
		if val > max {
			max = val
		}
	}

	return max
}
func DailyCalculateUserLevel(db *sql.DB, u model.RUser) {
	//now we calculate the remaining one
	av, users, err := GetLevelAV(db, u)
	if err != nil {
		utils.Log("Has Problem")
	}

	achievedLevel := "无星"
	achieve_level := 0

	rounds := config.ReturnActivity()
	for _, threshold := range rounds {
		fmt.Printf("High Checking Level: %s, Threshold: %d for user, reached %d numb %d \n", threshold.Level, threshold.Threshold, u.ID, av)

		if av >= (threshold.Threshold) {
			achievedLevel = threshold.Level
			achieve_level = config.PackageValue[threshold.Threshold]

		}
	}
	fmt.Printf("user %d reached level %s cause of av %d\n", u.ID, achievedLevel, av)

	if achievedLevel == "无星" {
		utils.Log("User %d still 无星 - AV %d too low", u.ID, av)
		return
	}

	//now we also check the lenth of users
	if len(users) <= 2 {
		utils.Log("%d at least 2 users", u.ID)
		return
	}
	nu := MaxSplitSum(users)
	numb := av - nu
	//numb = numb - u.AV
	reachedLevel := ""
	level_variable := 0
	thresholds := config.ReturnActivity()
	for _, threshold := range thresholds {
		fmt.Printf("Checking Level: %s, Threshold: %d for user, reached %d numb %d \n", threshold.Level, threshold.Threshold/2, u.ID, numb)

		if numb >= (threshold.Threshold / 2) {
			reachedLevel = threshold.Level
			level_variable = config.PackageValue[threshold.Threshold]
		}
	}
	fmt.Printf("user %d reached level %s cause of av %d\n and level variable is %d\n", u.ID, reachedLevel, numb, level_variable)

	if reachedLevel == "" {
		utils.Log("$d level reached still empty", u.ID)
		return
	}
	_, package_amount, fail := GetTopActivePackage(db, u.ID)
	if fail != nil {
		utils.Log("%d no package available", u.ID)
		return
	}
	fmt.Printf("user %d, reached level %s, achieved level %s, is already on %d, highest package needed is %d but user has %d\n",
		u.ID, achievedLevel, reachedLevel, u.Level, config.PackageInfo[achievedLevel], package_amount)
	// getIndex := GetLevelIndex(config.DailyActivity, achievedLevel)
	// getReached := GetLevelIndex(config.DailyHalf, reachedLevel)
	user_level := u.Level
	small_level := 0
	if achieve_level >= 1 && level_variable >= 1 && package_amount >= config.PackageInfo["1星"] {
		small_level = 1
	}
	if achieve_level >= 2 && level_variable >= 2 && package_amount >= config.PackageInfo["2星"] {
		small_level = 2
	}
	if achieve_level >= 3 && level_variable >= 3 && package_amount >= config.PackageInfo["3星"] {
		small_level = 3
	}
	if achieve_level >= 4 && level_variable >= 4 && package_amount >= config.PackageInfo["4星"] {
		small_level = 4
	}
	if achieve_level >= 5 && level_variable >= 5 && package_amount >= config.PackageInfo["5星"] {
		small_level = 5
	}
	if achieve_level >= 6 && level_variable >= 6 && package_amount >= config.PackageInfo["5星"] {
		small_level = 6
	}
	if achieve_level >= 7 && level_variable >= 7 && package_amount >= config.PackageInfo["6星"] {
		small_level = 7
	}

	fmt.Printf("user %d small level is %d while level vairbale is %d\n", u.ID, small_level, level_variable)

	if small_level > u.Level {
		_, err := db.Exec("UPDATE users SET level = ? WHERE id = ?", small_level, u.ID)
		if err != nil {
			utils.Log("⚠️ Failed to update user %d to level %d: %v", u.ID, small_level, err)
		} else {
			utils.Log("✅ User %d upgraded to level %d successfully", u.ID, small_level)
		}

		// _, err = db2.Exec("UPDATE users SET level = ? WHERE id = ?", small_level, u.ID)
		// if err != nil {
		// 	utils.Log("⚠️ Failed to update user %d to level %d: %v", u.ID, small_level, err)
		// }
	} else {
		utils.Log("❌ User %d already has level %d, no update needed", u.ID, small_level)
	}
	// for amount, level := range config.PackageOrder {
	// 	if getIndex >= (level-1) && getReached >= (level-1) {
	// 		if u.Level >= level {
	// 			utils.Log("❌ User %d already has level %d, no update needed", u.ID, level)
	// 		} else {
	// 			user_level = level
	// 			_, err := db.Exec("UPDATE users SET level = ? WHERE id = ?", level, u.ID)
	// 			if err != nil {
	// 				utils.Log("⚠️ Failed to update user %d to level %d: %v", u.ID, level, err)
	// 			} else {
	// 				utils.Log("✅ User %d upgraded to level %d successfully", u.ID, level)
	// 			}
	// 		}
	// 	}
	// }
	if user_level < 1 {
		utils.Log("%d level is 0", u.ID)
		return
	}
}

func GetLevelAV(db *sql.DB, user model.RUser) (int, []int, error) {
	rows, err := db.Query("SELECT av, group_av FROM users WHERE invited = ?", user.ID)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var total int
	var pl []int

	for rows.Next() {
		var avStr, groupAVStr sql.NullString
		if err := rows.Scan(&avStr, &groupAVStr); err != nil {
			return 0, nil, err
		}

		av := safeStrToInt(avStr.String)
		groupAV := safeStrToInt(groupAVStr.String)

		combined := av + groupAV
		total += combined
		pl = append(pl, combined)
	}
	pl = append(pl)
	return total, pl, nil
}
func safeStrToInt(s string) int {
	if s == "" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func GetLevelIndex(levels map[string]int, target string) int {
	keys := make([]string, 0, len(levels))
	for k := range levels {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return levels[keys[i]] < levels[keys[j]]
	})

	for i, k := range keys {
		if k == target {
			return i
		}
	}

	return -1 // not found
}

func GetTopActivePackage(db *sql.DB, userID int) (model.PackageLog, int, error) {
	query := `
		SELECT id, user_id, package_id, amount, add_time, frozen, active, day, already_added, method
FROM package_logs
WHERE user_id = ? AND active = 1
ORDER BY CAST(amount AS DECIMAL(20,2)) DESC
LIMIT 1;

	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return model.PackageLog{}, 0, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var topPackage model.PackageLog
	var topAmount float64
	found := false
	v := 0

	for rows.Next() {
		var pkg model.PackageLog
		if err := rows.Scan(
			&pkg.ID, &pkg.UserID, &pkg.PackageID, &pkg.Amount,
			&pkg.AddTime, &pkg.Frozen, &pkg.Active, &pkg.Day,
			&pkg.AlreadyAdded, &pkg.Method,
		); err != nil {
			continue
		}

		amountVal, err := strconv.ParseFloat(pkg.Amount, 64)
		if err != nil {
			continue
		}
		v = int(amountVal)

		if !found || amountVal > topAmount {
			topAmount = amountVal
			topPackage = pkg
			found = true
		}
	}

	if !found {
		return model.PackageLog{}, 0, fmt.Errorf("用户 [%d] -> 没有任务包", userID)
	}

	return topPackage, v, nil
}
