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

	err = database.AutoMigrate(&Policy{})
	if err != nil {
		panic("Failed to perform migration for `Policies`!")
	}

	err = database.AutoMigrate(&Organisation{})

	if err != nil {
		panic("Failed to perform migration for `Organisations`!")
	}

	err = database.AutoMigrate(&Repo{})

	if err != nil {
		panic("Failed to perform migration for `Repos`!")
	}

	err = database.AutoMigrate(&Project{})

	if err != nil {
		panic("Failed to perform migration for `Projects`!")
	}

	err = database.AutoMigrate(&Token{})

	if err != nil {
		panic("Failed to perform migration for `Tokens`!")
	}

	err = database.AutoMigrate(&User{})

	if err != nil {
		panic("Failed to perform migration for `Users`!")
	}

	err = database.AutoMigrate(&ProjectRun{})

	if err != nil {
		panic("Failed to perform migration for `ProjectRun`!")
	}

	err = database.AutoMigrate(&GithubAppInstallation{})

	if err != nil {
		panic("Failed to perform migration for `GithubAppInstallation`!")
	}

	err = database.AutoMigrate(&GithubApp{})

	if err != nil {
		panic("Failed to perform migration for `GithubApp`!")
	}

	err = database.AutoMigrate(&GithubAppInstallationLink{})

	if err != nil {
		panic("Failed to perform migration for `GithubAppInstallationLink`!")
	}

	DB = database
}
