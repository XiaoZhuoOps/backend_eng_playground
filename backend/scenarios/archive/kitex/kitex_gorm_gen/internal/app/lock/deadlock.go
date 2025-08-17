package web

import (
	"errors"
	"gorm.io/gorm"
	"strings"
	"time"
)

// Retry mechanism for deadlocks
func updateUserEmail(db *gorm.DB, userID uint, newEmail string) error {
	for i := 0; i < 3; i++ { // Retry up to 3 times
		err := db.Transaction(func(tx *gorm.DB) error {
			var user User
			if err := tx.Clauses(gorm.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
				return err
			}

			user.Email = newEmail
			if err := tx.Save(&user).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			// Check if it's a deadlock error
			if isDeadlockError(err) {
				time.Sleep(time.Millisecond * 100) // Wait before retrying
				continue
			}
			return err
		}

		return nil
	}
	return errors.New("could not complete transaction due to deadlocks")
}

func isDeadlockError(err error) bool {
	return strings.Contains(err.Error(), "deadlock")
}
