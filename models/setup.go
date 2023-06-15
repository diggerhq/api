package models

import (
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

var DB *gorm.DB

func ConnectDatabase() {

	database, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic("Failed to connect to database!")
	}

	err = database.AutoMigrate(&Test{})
	if err != nil {
		panic("Failed to perform migration for `Test`!")
	}

	err = database.AutoMigrate(&Policy{})
	if err != nil {
		panic("Failed to perform migration for `Policies`!")
	}

	DB = database
}
