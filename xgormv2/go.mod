module github.com/Aoi-hosizora/ahlib-db/xgormv2

go 1.15

require (
	github.com/Aoi-hosizora/ahlib v0.0.0-00010101000000-000000000000
	github.com/VividCortex/mysqlerr v1.0.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/mattn/go-sqlite3 v1.14.9
	github.com/sirupsen/logrus v1.8.1
	gorm.io/driver/mysql v1.2.3
	gorm.io/driver/sqlite v1.2.6
	gorm.io/gorm v1.22.4
)

replace github.com/Aoi-hosizora/ahlib => ../../ahlib
