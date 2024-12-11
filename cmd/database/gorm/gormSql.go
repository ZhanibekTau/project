package gormSql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"project/cmd/database/model"
)

// DBConfig represents db configuration
type DBConfig struct {
	Host     string
	Port     int
	User     string
	DBName   string
	Password string
}

func DbURL(dbConfig *DBConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)
}

func NewGormSqlDB(dbConfig *DBConfig) (*gorm.DB, error) {
	Database, err := gorm.Open("mysql", DbURL(dbConfig))
	if err != nil {
		fmt.Println("Status:", err)
		return nil, err
	}

	err = Database.DB().Ping()
	if err != nil {
		return nil, err
	}
	Database.AutoMigrate(&model.User{})

	return Database, nil
}
