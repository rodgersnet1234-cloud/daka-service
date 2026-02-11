package handlers

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
)

func CalculateLevel(dbv *sql.DB, c *gin.Context) {
	// phone := c.Query("phone")
	//db.AppendPhoneNumber(phone)
	// functions.HandleTask(dbv, phone)
}

func IncrementGroupCounts(db *sql.DB, currentUserID int, amount float64) error {
	depth := 0
	fmt.Println("reached here -> start")
	// logger := log.New(logFile, "", log.LstdFlags)
	fmt.Printf("reached here -> entering %d and amount is %v \n", currentUserID, amount)
	var recurse func(userID int, amt float64, depth int) error
	recurse = func(userID int, amt float64, depth int) error {
		depth++
		fmt.Printf("🔍 Checking inviter for user ID: %d\n", userID)
		// depth := 1
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
			InsertGroup(db, currentUserID, int(parentID.Int64), depth)
			fmt.Printf("reached here , parent id is %d and depth is %d\n", parentID.Int64, depth)
			return recurse(int(parentID.Int64), amt, depth)
		}
		UpdateUser(db, currentUserID)
		fmt.Printf("🚫 User %d has no inviter. Ending recursion.\n", userID)
		return nil
	}
	err := recurse(currentUserID, amount, depth)
	if err != nil {
		fmt.Printf("❌ Error during recursion: %v\n", err)
		return err
	}
	fmt.Printf("🚀 Starting incrementGroupCounts for user %d amount %.2f\n", currentUserID, amount)
	return nil
}

func InsertGroup(db *sql.DB, userID int, parent int, depth int) error {
	_, err := db.Exec("INSERT INTO user_groups (user_id, ancestor_id, depth) VALUES (?, ?, ?)", userID, parent, depth)
	return err
}

func UpdateUser(db *sql.DB, userId int) error {
	_, err := db.Exec("UPDATE users SET added = 1 WHERE id = ?", userId)
	return err
}
