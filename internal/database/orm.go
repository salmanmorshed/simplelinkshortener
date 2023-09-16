package database

import (
	"fmt"

	"github.com/salmanmorshed/simplelinkshortener/internal/config"
	"github.com/salmanmorshed/simplelinkshortener/internal/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func CreateGORM(conf *config.AppConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	if conf.Database.Type == "postgresql" {
		db, err = gorm.Open(postgres.Open(fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s%s",
			conf.Database.Host,
			conf.Database.Port,
			conf.Database.Username,
			conf.Database.Password,
			conf.Database.Name,
			utils.StringifyConfigDBExtraArgs(conf),
		)), &gorm.Config{})
	} else if conf.Database.Type == "mysql" {
		db, err = gorm.Open(mysql.Open(fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s%s",
			conf.Database.Username,
			conf.Database.Password,
			conf.Database.Host,
			conf.Database.Port,
			conf.Database.Name,
			utils.StringifyConfigDBExtraArgs(conf),
		)), &gorm.Config{})
	} else if conf.Database.Type == "sqlite" {
		db, err = gorm.Open(sqlite.Open(conf.Database.Name), &gorm.Config{})
	} else {
		err = fmt.Errorf("unsupported database type: %v", conf.Database.Type)
	}
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&User{}, &Link{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
