package functions

import (
	"daka/config"
	"daka/model"
	"daka/utils"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

func ProcessPackageLog(db *sql.DB, pkg model.PackageTask, user model.RUser, date string) (error, string) {
	utils.Log("➡️ Starting ProcessPackageLog for userID=%d pkgLogID=%d date=%s", user.ID, pkg.ID, date)
	mainUser, err := GetUserByID(db, user.ID)
	if err != nil {
		utils.Log("❌ GetUserByID failed for inviter %d: %v", mainUser.ID, err)
		return err, "GetUserByID"
	}
	pkgLog, err := GetPackageLogByID(db, pkg.ID)
	if err != nil {
		utils.Log("❌ Failed to get PackageLog by ID %d: %v", pkg.ID, err)
		return err, "PackageLog"
	}
	utils.Log("📦 PackageLog retrieved: %+v", pkgLog)

	pkgInfo, err := GetPackageByID(db, pkgLog.PackageID)
	if err != nil {
		utils.Log("❌ Failed to get Package by ID %d: %v", pkgLog.PackageID, err)
		return err, "PackageId"
	}
	utils.Log("📦 Package info retrieved: %+v", pkgInfo)

	trans := "12312312" //RandomString(10)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc).Format("2006-01-02 15:04:05")
	// now := "2025-10-21"
	// Daily earnings
	dailyReceive := Round(pkgInfo.TradableAmount/float64(pkgLog.Days), 2)
	utils.Log("💰 Calculated dailyReceive: %.4f (tradable_amount: %.4f / 30)", dailyReceive, pkgInfo.TradableAmount)

	prevBalance, err := strconv.ParseFloat(mainUser.Shouyi, 64)
	if err != nil {
		utils.Log("⚠️ User %d shouyi parse error: %v (value: %s)", user.ID, err, user.Shouyi)
		prevBalance = 0
	}
	newShouyi := prevBalance + dailyReceive
	utils.Log("⬆️ User %d shouyi: %.4f -> %.4f", user.ID, prevBalance, newShouyi)

	_, err = InsertShouyiLog(db, user.ID, dailyReceive, now, trans, prevBalance, newShouyi, pkg.ID, config.ShortPackageLevel[pkgInfo.PackageValue])
	if err != nil {
		utils.Log("❌ InsertShouyiLog failed for user %d: %v", user.ID, err)
		return err, "InsertShouyiLog"
	}
	// err = InsertShouyiLog2(db2, int(shouyiId), user.ID, dailyReceive, now, trans, prevBalance, newShouyi, pkg.ID, config.ShortPackageLevel[pkgInfo.PackageValue])
	// if err != nil {
	// 	utils.Log("❌ InsertShouyiLog failed for user %d: %v", user.ID, err)
	// 	return err, "InsertShouyiLog2"
	// }
	utils.Log("✅ Inserted ShouyiLog for user %d", user.ID)

	_, err = UpdateUserShouyi(db, user.ID, dailyReceive)
	if err != nil {
		utils.Log("❌ UpdateUserShouyi failed for user %d: %v", user.ID, err)
		return err, "UpdateUserShouyi"
	}
	utils.Log("✅ Updated User Shouyi for user %d by %.4f", user.ID, dailyReceive)
	// err = UpdateUserShouyi2(db2, user.ID, shouyiBalance)
	// if err != nil {
	// 	utils.Log("❌ UpdateUserShouyi failed for user %d: %v", user.ID, err)
	// 	return err, "UpdateUserShouyi2"
	// }
	// Jifen
	newDailyJifen := Round(pkgInfo.TotalReward/float64(pkgLog.Days), 2)
	dailyJifen := newDailyJifen - dailyReceive
	utils.Log("💎 Calculated dailyJifen: %.4f (daily_reward: %.4f - dailyReceive: %.4f)", dailyJifen, pkgInfo.DailyReward, dailyReceive)

	prevJifen, err := strconv.ParseFloat(mainUser.Jifen, 64)
	if err != nil {
		utils.Log("⚠️ User %d jifen parse error: %v (value: %s)", user.ID, err, user.Jifen)
		prevJifen = 0
	}
	newJifen := prevJifen + dailyJifen
	utils.Log("⬆️ User %d jifen: %.4f -> %.4f", user.ID, prevJifen, newJifen)

	_, err = InsertJifenLog(db, user.ID, dailyJifen, now, config.ShortPackageLevel[pkgInfo.PackageValue], prevJifen, newJifen, trans)
	if err != nil {
		utils.Log("❌ InsertJifenLog failed for user %d: %v", user.ID, err)
		return err, "InsertJifenLog"
	}
	// InsertJifenLog2(db2, int(jifenId), user.ID, dailyJifen, now, config.ShortPackageLevel[pkgInfo.PackageValue], prevJifen, newJifen, trans)
	// utils.Log("✅ Inserted JifenLog2 for user %d", user.ID)
	// if err != nil {
	// 	utils.Log("❌ InsertJifenLog2 failed for user %d: %v", user.ID, err)
	// 	return err, "InsertJifenLog2"
	// }
	_, err = UpdateUserJifen(db, user.ID, dailyJifen)
	if err != nil {
		utils.Log("❌ UpdateUserJifen failed for user %d: %v", user.ID, err)
		return err, "UpdateUserJifen"
	}

	// err = UpdateUserJifen2(db2, user.ID, jifenBalance)
	// if err != nil {
	// 	utils.Log("❌ UpdateUserJifen failed for user %d: %v", user.ID, err)
	// 	return err, "UpdateUserJifen"
	// }
	utils.Log("✅ Updated User Jifen for user %d by %.4f", user.ID, dailyJifen)

	// Update package log day
	err = UpdatePackageLogDay(db, pkg.ID)
	if err != nil {
		utils.Log("❌ UpdatePackageLogDay failed for packageLogID %d: %v", pkg.ID, err)
		return err, "UpdatePackageLogDay"
	}

	// err = UpdatePackageLogDay(db2, pkg.ID)
	// if err != nil {
	// 	utils.Log("❌ UpdatePackageLogDay failed for packageLogID %d: %v", pkg.ID, err)
	// 	return err, "UpdatePackageLogDay"
	// }
	utils.Log("✅ Updated PackageLog day for packageLogID %d", pkg.ID)

	amount, err := strconv.ParseFloat(pkgLog.Amount, 64)
	if err != nil {
		utils.Log("⚠️ PackageLog amount parse error for user %d: %v (value: %s)", user.ID, err, pkgLog.Amount)
		amount = 0
	}
	avAmount := Round(amount/10, 0)
	utils.Log("📊 Calculated AV amount: %.0f (amount %.4f / 10)", avAmount, amount)

	// err = UpdateUserAV(db, user.ID, avAmount)
	// if err != nil {
	// 	utils.Log("❌ UpdateUserAV failed for user %d: %v", user.ID, err)
	// 	return err, "UpdateUserAV"
	// }
	// utils.Log("✅ Updated User AV and total for user %d by %.0f", user.ID, avAmount)

	// Handle inviter earnings
	if user.Invited.Valid {
		inviterID := int(user.Invited.Int64)
		utils.Log("👥 User %d has inviter: %d", user.ID, inviterID)

		inviter, err := GetUserByID(db, inviterID)
		if err != nil {
			utils.Log("❌ GetUserByID failed for inviter %d: %v", inviterID, err)
			return err, "GetUserByID"
		}
		utils.Log("👤 Inviter info: %+v", inviter)

		packageValue, err := strconv.ParseInt(pkgInfo.PackageValue, 10, 64)
		if err != nil {
			utils.Log("⚠️ Error parsing package value '%s': %v", pkgInfo.PackageValue, err)
			packageValue = 0
		}

		inviterShouyi, ok := config.DailyPackage[int(packageValue)]
		if !ok {
			utils.Log("⚠️ DailyPackage entry not found for package value: %d", packageValue)
			inviterShouyi = 0
		}
		utils.Log("💸 Calculated inviter shouyi: %.4f for packageValue %d", inviterShouyi, packageValue)

		prevInviterBalance, err := strconv.ParseFloat(inviter.Shouyi, 64)
		if err != nil {
			utils.Log("⚠️ Inviter %d shouyi parse error: %v (value: %s)", inviter.ID, err, inviter.Shouyi)
			prevInviterBalance = 0
		}
		newInviterBalance := prevInviterBalance + inviterShouyi
		utils.Log("⬆️ Inviter %d shouyi: %.4f -> %.4f", inviter.ID, prevInviterBalance, newInviterBalance)

		source := fmt.Sprintf("QY%d的%s", user.ID, config.ShortPackageLevel[pkgInfo.PackageValue])
		_, err = UpdateUserShouyi(db, inviter.ID, inviterShouyi)
		if err != nil {
			utils.Log("❌ UpdateUserInviterEarnings failed for user %d: %v", user.ID, err)
			return err, "UpdateUserInviterEarnings"
		}
		utils.Log("✅ Updated User Shouyi for user %d by %.4f", user.ID, inviterShouyi)
		// err = UpdateUserShouyi2(db2, inviter.ID, shouyiBalance)
		// if err != nil {
		// 	utils.Log("❌ UpdateUserInviterEarnings2 failed for user %d: %v", user.ID, err)
		// 	return err, "UpdateUserInviterEarnings2"
		// }
		//err = UpdateUserInviterEarnings(db, inviter.ID, inviterShouyi, avAmount)
		//if err != nil {
		//	utils.Log("❌ UpdateUserInviterEarnings failed for inviter %d: %v", inviter.ID, err)
		//	return err, "UpdateUserInviterEarnings"
		//}
		utils.Log("✅ Updated inviter earnings for inviter %d by %.4f (shouyi) and %.0f (AV)", inviter.ID, inviterShouyi, avAmount)

		_, err = InsertShouyiLog(db, inviter.ID, inviterShouyi, now, trans, prevInviterBalance, newInviterBalance, pkg.ID, source)
		if err != nil {
			utils.Log("❌ InsertShouyiLog failed for inviter %d: %v", inviter.ID, err)
			return err, "InsertShouyiLog"
		}
		// err = InsertShouyiLog2(db2, int(inviterShouyiId), inviter.ID, inviterShouyi, now, trans, prevInviterBalance, newInviterBalance, pkg.ID, source)
		// if err != nil {
		// 	utils.Log("❌ InsertShouyiLog failed for inviter %d: %v", inviter.ID, err)
		// 	return err, "InsertShouyiLog"
		// }
		utils.Log("✅ Inserted ShouyiLog for inviter %d", inviter.ID)

		_, err = InsertPackageActiveLog(db, user.ID, pkg.ID, date)
		if err != nil {
			utils.Log("❌ InsertPackageActiveLog failed for user %d: %v", user.ID, err)
			return err, "InsertPackageActiveLog"
		}

		// err = InsertPackageActiveLog2(db2, int(packageId), user.ID, pkg.ID, date)
		// if err != nil {
		// 	utils.Log("❌ InsertPackageActiveLog failed for user %d: %v", user.ID, err)
		// 	return err, "InsertPackageActiveLog"
		// }
		utils.Log("✅ Inserted PackageActiveLog for user %d", user.ID)

	} else {
		utils.Log("👥 User %d has no inviter", user.ID)
	}

	//now we also update activity
	new_day := pkgLog.Day + 1
	if new_day >= pkgLog.Days {
		updatePackageActivity(db, pkgLog)
		// updatePackageActivity(db2, pkgLog)
	}
	_, err = InsertUserActive(db, user.ID, avAmount, date, pkg.ID)
	if err != nil {
		utils.Log("❌ InsertUserActive failed for user %d: %v", user.ID, err)
		return err, "InsertUserActive"
	}
	// err = InsertUserActive2(db2, int(activeId), user.ID, avAmount, date, pkg.ID)
	// if err != nil {
	// 	utils.Log("❌ InsertUserActive failed for user %d: %v", user.ID, err)
	// 	return err, "InsertUserActive2"
	// }
	utils.Log("✅ Inserted UserActive record for user %d", user.ID)

	utils.Log("✅ ProcessPackageLog completed successfully for user %d", user.ID)
	return nil, ""
}

func updatePackageActivity(db *sql.DB, pkgLog model.PackageLog) error {
	_, err := db.Exec("UPDATE package_logs SET active = 2  WHERE id = ?", pkgLog.ID)
	return err
}

func DecrementGroupCounts(db *sql.DB, currentUserID int, amount float64) error {
	fmt.Println("reached here -> start")
	logFile, err := os.OpenFile("group_av6.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	fmt.Printf("reached here -> entering %d and amount is %v \n", currentUserID, amount)
	var recurse func(userID int, amt float64) error
	recurse = func(userID int, amt float64) error {
		fmt.Printf("🔍 Checking inviter for user ID: %d\n", userID)

		var parentID sql.NullInt64
		err := db.QueryRow("SELECT invited FROM users WHERE id = ?", userID).Scan(&parentID)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("⚠️ No user found with ID: %d. Ending recursion.\n", userID)
				return nil
			}
			fmt.Printf("❌ Error querying inviter for user %d: %v\n", userID, err)
			return fmt.Errorf("failed to query invited for user %d: %w", userID, err)
		}

		if parentID.Valid {
			fmt.Printf("👥 User %d invited by %d. Updating group_av by %.2f\n", userID, parentID.Int64, amt)
			_, err := db.Exec("UPDATE users SET group_av = group_av - ? WHERE id = ?", amt, parentID.Int64)
			if err != nil {
				fmt.Printf("❌ Failed to update group_av for user %d: %v\n", parentID.Int64, err)
				return fmt.Errorf("failed to update group_av for user %d: %w", parentID.Int64, err)
			}
			fmt.Printf("✅ Successfully updated group_av for user %d\n", parentID.Int64)

			loc, _ := time.LoadLocation("Asia/Shanghai")
			timestamp := time.Now().In(loc).Format("2006-01-02 15:04:05")

			fmt.Printf("[%s] Logged group_av update: +%.2f for user ID: %d (invited by %d)\n",
				timestamp, amt, parentID.Int64, userID)

			// Recurse up the inviter chain
			fmt.Printf("🔄 Recursing for inviter %d\n", parentID.Int64)
			return recurse(int(parentID.Int64), amt)
		}

		fmt.Printf("🚫 User %d has no inviter. Ending recursion.\n", userID)
		return nil
	}

	fmt.Printf("🚀 Starting decrenent for user %d amount %.2f\n", currentUserID, amount)
	err = recurse(currentUserID, amount)
	if err != nil {
		logger.Printf("❌ Error during recursion: %v\n", err)
		return err
	}
	fmt.Printf("🏁 Finished incrementGroupCounts for user %d\n", currentUserID)
	return nil
}

func incrementGroupCounts(db *sql.DB, currentUserID int, amount float64) error {
	fmt.Println("reached here -> start")
	logFile, err := os.OpenFile("group_av5.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	fmt.Printf("reached here -> entering %d and amount is %v \n", currentUserID, amount)
	var recurse func(userID int, amt float64) error
	recurse = func(userID int, amt float64) error {
		fmt.Printf("🔍 Checking inviter for user ID: %d\n", userID)

		var parentID sql.NullInt64
		err := db.QueryRow("SELECT invited FROM users WHERE id = ?", userID).Scan(&parentID)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Printf("⚠️ No user found with ID: %d. Ending recursion.\n", userID)
				return nil
			}
			fmt.Printf("❌ Error querying inviter for user %d: %v\n", userID, err)
			return fmt.Errorf("failed to query invited for user %d: %w", userID, err)
		}

		if parentID.Valid {
			fmt.Printf("👥 User %d invited by %d. Updating group_av by %.2f\n", userID, parentID.Int64, amt)
			_, err := db.Exec("UPDATE users SET group_av = group_av + ? WHERE id = ?", amt, parentID.Int64)
			if err != nil {
				fmt.Printf("❌ Failed to update group_av for user %d: %v\n", parentID.Int64, err)
				return fmt.Errorf("failed to update group_av for user %d: %w", parentID.Int64, err)
			}
			fmt.Printf("✅ Successfully updated group_av for user %d\n", parentID.Int64)
			var newJifen string
			err = db.QueryRow(
				"SELECT group_av FROM users WHERE id = ?",
				parentID.Int64,
			).Scan(&newJifen)
			// _, err = db2.Exec("UPDATE users SET group_av =  ? WHERE id = ?", newJifen, parentID.Int64)
			// if err != nil {
			// 	fmt.Printf("❌ Failed to update group_av for user %d: %v\n", parentID.Int64, err)
			// 	return fmt.Errorf("failed to update group_av for user %d: %w", parentID.Int64, err)
			// }
			loc, _ := time.LoadLocation("Asia/Shanghai")
			timestamp := time.Now().In(loc).Format("2006-01-02 15:04:05")

			fmt.Printf("[%s] Logged group_av update: +%.2f for user ID: %d (invited by %d)\n",
				timestamp, amt, parentID.Int64, userID)

			// Recurse up the inviter chain
			fmt.Printf("🔄 Recursing for inviter %d\n", parentID.Int64)
			return recurse(int(parentID.Int64), amt)
		}

		fmt.Printf("🚫 User %d has no inviter. Ending recursion.\n", userID)
		return nil
	}

	fmt.Printf("🚀 Starting incrementGroupCounts for user %d amount %.2f\n", currentUserID, amount)
	err = recurse(currentUserID, amount)
	if err != nil {
		logger.Printf("❌ Error during recursion: %v\n", err)
		return err
	}
	fmt.Printf("🏁 Finished incrementGroupCounts for user %d\n", currentUserID)
	return nil
}

func GetPackageLogByID(db *sql.DB, id int) (model.PackageLog, error) {
	var log model.PackageLog
	err := db.QueryRow("SELECT id, user_id, package_id, amount, day, days FROM package_logs WHERE id = ?", id).
		Scan(&log.ID, &log.UserID, &log.PackageID, &log.Amount, &log.Day, &log.Days)
	return log, err
}

func GetPackageByID(db *sql.DB, id int) (model.Package, error) {
	var p model.Package
	err := db.QueryRow("SELECT id, tradable_amount, daily_reward, package_value, total_reward FROM packages WHERE id = ?", id).
		Scan(&p.ID, &p.TradableAmount, &p.DailyReward, &p.PackageValue, &p.TotalReward)
	return p, err
}

func InsertShouyiLog(db *sql.DB, userID int, amount float64, t, trans string, before, after float64, packageID int, source ...string) (int64, error) {
	sourceVal := ""
	if len(source) > 0 {
		sourceVal = source[0]
	}
	result, err := db.Exec("INSERT INTO shouyi_logs (user_id, amount, time, packages, trans, balance_before, balance_after, package_id, source) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		userID, amount, t, "[]", trans, before, after, packageID, sourceVal)
	if err != nil {
		return 0, err
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return insertID, nil
}
func InsertShouyiLog2(db *sql.DB, id int, userID int, amount float64, t, trans string, before, after float64, packageID int, source ...string) error {
	sourceVal := ""
	if len(source) > 0 {
		sourceVal = source[0]
	}
	_, err := db.Exec("INSERT INTO shouyi_logs (id,user_id, amount, time, packages, trans, balance_before, balance_after, package_id, source) VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, userID, amount, t, "[]", trans, before, after, packageID, sourceVal)
	return err
}

func ImportInsertShouyiLog2(db *sql.DB, id string, userID string, amount string, t, trans string, before, after string, packageID string, sourceVal string) error {
	_, err := db.Exec("INSERT INTO shouyi_logs (id,user_id, amount, time, packages, trans, balance_before, balance_after, package_id, source) VALUES (?,?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, userID, amount, t, "[]", trans, before, after, packageID, sourceVal)
	return err
}

func InsertJifenLog(db *sql.DB, userID int, amount float64, t, source string, before, after float64, trans string) (int64, error) {
	result, err := db.Exec("INSERT INTO jifen_log (user_id, amount, tt, source, balance_before, balance_after, trans) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, amount, t, source, before, after, trans)
	if err != nil {
		return 0, err
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return insertID, nil
}
func InsertJifenLog2(db *sql.DB, id int, userID int, amount float64, t, source string, before, after float64, trans string) error {
	_, err := db.Exec("INSERT INTO jifen_log (id,user_id, amount, tt, source, balance_before, balance_after, trans) VALUES (?,?, ?, ?, ?, ?, ?, ?)",
		id, userID, amount, t, source, before, after, trans)
	return err
}

func UpdateUserShouyi(db *sql.DB, userID int, amount float64) (string, error) {
	_, err := db.Exec("UPDATE users SET shouyi = shouyi + ? WHERE id = ?", amount, userID)

	if err != nil {
		return "0", err
	}

	// 2. Query updated value
	var newJifen string
	err = db.QueryRow(
		"SELECT shouyi FROM users WHERE id = ?",
		userID,
	).Scan(&newJifen)

	if err != nil {
		return "0", err
	}

	return newJifen, nil
}

func UpdateUserJifen(db *sql.DB, userID int, amount float64) (string, error) {
	// 1. Update
	_, err := db.Exec(
		"UPDATE users SET jifen = jifen + ? WHERE id = ?",
		amount, userID,
	)
	if err != nil {
		return "0", err
	}

	// 2. Query updated value
	var newJifen string
	err = db.QueryRow(
		"SELECT jifen FROM users WHERE id = ?",
		userID,
	).Scan(&newJifen)

	if err != nil {
		return "0", err
	}

	return newJifen, nil
}

func UpdateUserShouyi2(db *sql.DB, userID int, amount string) error {
	_, err := db.Exec("UPDATE users SET shouyi =  ? WHERE id = ?", amount, userID)
	return err
}
func UpdateUserJifen2(db *sql.DB, userID int, amount string) error {
	_, err := db.Exec("UPDATE users SET jifen =  ? WHERE id = ?", amount, userID)
	return err
}

func UpdatePackageLogDay(db *sql.DB, id int) error {
	_, err := db.Exec("UPDATE package_logs SET day = day + 1 WHERE id = ?", id)
	return err
}
func UpdatePackageLogDay2(db *sql.DB, id int) error {
	_, err := db.Exec("UPDATE package_logs SET day = day + 1 WHERE id = ?", id)
	return err
}

func UpdateUserAV3(db *sql.DB, userID int, avAmount string) error {
	_, err := db.Exec("UPDATE users SET av = ? WHERE id = ?", avAmount, userID)
	return err
}

func UpdateUserAV(db *sql.DB, userID int, avAmount float64) (string, error) {
	_, err := db.Exec("UPDATE users SET av = av + ?, total = total + ? WHERE id = ?", avAmount, avAmount, userID)
	if err != nil {
		return "0", err
	}

	// 2. Query updated value
	var newJifen string
	err = db.QueryRow(
		"SELECT av FROM users WHERE id = ?",
		userID,
	).Scan(&newJifen)

	if err != nil {
		return "0", err
	}

	return newJifen, nil
}

func UpdateUserAV1(db *sql.DB, userID int, avAmount float64) error {
	_, err := db.Exec("UPDATE users SET av =  ?, total = total + ? WHERE id = ?", avAmount, avAmount, userID)
	return err
}

func UpdateUserAV2(db *sql.DB, userID int, avAmount float64) error {
	_, err := db.Exec("UPDATE users SET av = av - ?, total = total - ? WHERE id = ?", avAmount, avAmount, userID)
	return err
}

func InsertUserActive(db *sql.DB, userID int, amount float64, day string, packageLogID int) (int64, error) {
	result, err := db.Exec("INSERT INTO user_active (user_id, amount, day, package_logs_id) VALUES (?, ?, ?, ?)", userID, amount, day, packageLogID)
	if err != nil {
		return 0, err
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return insertID, nil
}

func InsertUserActive2(db *sql.DB, id int, userID int, amount float64, day string, packageLogID int) error {
	_, err := db.Exec("INSERT INTO user_active (id,user_id, amount, day, package_logs_id) VALUES (?,?, ?, ?, ?)", id, userID, amount, day, packageLogID)
	return err
}

func UpdateUserInviterEarnings(db *sql.DB, inviterID int, shouyiAmount, avAmount float64) error {
	_, err := db.Exec("UPDATE users SET shouyi = shouyi + ?, total_group = total_group + ? WHERE id = ?", shouyiAmount, avAmount, inviterID)
	return err
}

func InsertPackageActiveLog(db *sql.DB, userID, packageLogID int, day string) (int64, error) {
	result, err := db.Exec("INSERT INTO package_active_log (user_id, package_logs_id, day) VALUES (?, ?, ?)", userID, packageLogID, day)
	if err != nil {
		return 0, err
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return insertID, nil
}

func InsertPackageActiveLog2(db *sql.DB, id int, userID, packageLogID int, day string) error {
	_, err := db.Exec("INSERT INTO package_active_log (id,user_id, package_logs_id, day) VALUES (?,?, ?, ?)", id, userID, packageLogID, day)
	return err
}

func UpdateDownGroupCounts(userID int, db *sql.DB, amount float64) {
	// Placeholder for your recursive or delegated logic
	fmt.Printf("Updating down group counts for user %d with +%.2f\n", userID, amount)
	err := incrementGroupCounts(db, userID, amount)
	if err != nil {
		log.Printf("❌ Error updating down group counts for user %d: %v", userID, err)
	}
}

func Round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// func RandomString(n int) string {
// 	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
// 	s := make([]rune, n)
// 	for i := range s {
// 		s[i] = letters[rand.Intn(len(letters))]
// 	}
// 	return string(s)
// }
