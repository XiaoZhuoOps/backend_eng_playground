package repo

import (
	"backend_eng_playground/pkg/repo/config"
	"backend_eng_playground/pkg/repo/model/query"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(mysql.Open(config.DSN), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.Default.LogMode(logger.Info), // 这里会显示实际的SQL
	})
	if err != nil {
		panic(err)
	}
	query.SetDefault(DB)
}
