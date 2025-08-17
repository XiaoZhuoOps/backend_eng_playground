package web

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

/*
实现一个upsert接口，外部传入一个对象，包含唯一key，根据数据库中是否存在该记录进行更新和删除
*/
// CREATE TABLE my_entities (
// id INT AUTO_INCREMENT PRIMARY KEY,
// external_id VARCHAR(255) NOT NULL,
// params JSON NOT NULL,
// created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
// );

// -- Enforcing uniqueness of external_id
// ALTER TABLE my_entities
// ADD UNIQUE KEY (external_id);

// Solution1 Directly Insert With Unique Key, return error if it is existing record
package main

import (
"fmt"
"gorm.io/driver/mysql"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

type MyEntity struct {
	ID          uint           `gorm:"primaryKey"`
	ExternalID  string         `gorm:"uniqueIndex"` // Enforce uniqueness
	Params      JSON           `gorm:"type:json"`   // Custom type or a string for JSON
	CreatedAt   Time           // GORM will handle timestamps automatically
	UpdatedAt   Time
}

func main() {
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Auto-migrate schema
	db.AutoMigrate(&MyEntity{})

	// Create
	entity := MyEntity{
		ExternalID: "ext_12345",
		Params:     JSON(`{"key": "value"}`),
	}
	// Insert record - if ExternalID is unique, it succeeds; otherwise, an error is returned
	err = db.Create(&entity).Error
	if err != nil {
		fmt.Println("Create error:", err)
		// Handle uniqueness violation or other DB errors
	}
}

// Solution2 Upsert INSERT ... ON DUPLICATE KEY UPDATE
// GORM's "FirstOrCreate" can behave somewhat like an upsert
entity := MyEntity{
	ExternalID: "ext_12345",
	Params:     JSON(`{"key": "value"}`),
}

db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "external_id"}}, // unique constraint column
	DoUpdates: clause.AssignmentColumns([]string{"params", "updated_at"}),
}).Create(&entity)

// Solution3 Select For Update
func CreateEntityWithLock(db *gorm.DB, externalID string, params JSON) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// SELECT ... FOR UPDATE to lock potential existing row
		var existing MyEntity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("external_id = ?", externalID).
			First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Not found, insert new
				newEntity := MyEntity{
					ExternalID: externalID,
					Params:     params,
				}
				if err := tx.Create(&newEntity).Error; err != nil {
					return err
				}
			} else {
				// some other error
				return err
			}
		} else {
			// If it already exists, either fail or update
			// Example: Let's just fail to ensure only one creation
			return fmt.Errorf("entity with external_id %s already exists", externalID)
		}
		return nil
	})
}

// Solution4 Separate Lock or Registry Table 尽可能减少对主表的影响
CREATE TABLE external_id_registry (
external_id VARCHAR(255) NOT NULL PRIMARY KEY,
claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE my_entities (
id INT AUTO_INCREMENT PRIMARY KEY,
external_id VARCHAR(255) NOT NULL UNIQUE,
params JSON NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

func CreateEntityRegistryLock(db *gorm.DB, externalID string, params JSON) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Try to claim the external_id in the registry
		registry := struct {
			ExternalID string `gorm:"primaryKey"`
			ClaimedAt  time.Time
		}{
			ExternalID: externalID,
		}
		if err := tx.Create(&registry).Error; err != nil {
			// Duplicate primary key in registry indicates someone else claimed it
			return fmt.Errorf("external_id %s is already claimed", externalID)
		}

		// Now insert the entity
		newEntity := MyEntity{
			ExternalID: externalID,
			Params:     params,
		}
		if err := tx.Create(&newEntity).Error; err != nil {
			return err
		}

		return nil
	})
}
