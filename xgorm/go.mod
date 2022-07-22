module github.com/Aoi-hosizora/ahlib-db/xgorm

go 1.15

require (
	github.com/Aoi-hosizora/ahlib v0.0.0-00010101000000-000000000000
	github.com/Aoi-hosizora/ahlib-db/xdbutils v0.0.0-00010101000000-000000000000
	github.com/VividCortex/mysqlerr v1.0.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.16
	github.com/lib/pq v1.1.1
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/sirupsen/logrus v1.8.1
)

replace (
	github.com/Aoi-hosizora/ahlib => ../../ahlib
	github.com/Aoi-hosizora/ahlib-db/xdbutils => ./../xdbutils
)
