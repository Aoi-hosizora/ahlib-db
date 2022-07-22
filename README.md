# ahlib-db

[![Build Status](https://travis-ci.com/Aoi-hosizora/ahlib-db.svg?branch=master)](https://travis-ci.com/Aoi-hosizora/ahlib-db)
[![Codecov](https://codecov.io/gh/Aoi-hosizora/ahlib-db/branch/master/graph/badge.svg)](https://codecov.io/gh/Aoi-hosizora/ahlib-db)
[![Go Report Card](https://goreportcard.com/badge/github.com/Aoi-hosizora/ahlib-db)](https://goreportcard.com/report/github.com/Aoi-hosizora/ahlib-db)
[![License](http://img.shields.io/badge/license-mit-blue.svg)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/Aoi-hosizora/ahlib-db)](https://github.com/Aoi-hosizora/ahlib-db/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/Aoi-hosizora/ahlib-db.svg)](https://pkg.go.dev/github.com/Aoi-hosizora/ahlib-db)

+ A personal golang library for db development, including: mysql+sqlite+postgres ([gorm (v1)](https://github.com/jinzhu/gorm) / [gorm (v2)](https://github.com/go-gorm/gorm)), redis ([go-redis](https://github.com/go-redis/redis)), neo4j ([neo4j-go-driver (v1)](https://github.com/neo4j/neo4j-go-driver)), requires `Go >= 1.15`.

### Related libraries

+ [Aoi-hosizora/ahlib](https://github.com/Aoi-hosizora/ahlib)
+ [Aoi-hosizora/ahlib-more](https://github.com/Aoi-hosizora/ahlib-more)
+ [Aoi-hosizora/ahlib-web](https://github.com/Aoi-hosizora/ahlib-web)
+ [Aoi-hosizora/ahlib-db](https://github.com/Aoi-hosizora/ahlib-db)

### Packages

+ xdbutils/*
+ xgorm
+ xgormv2
+ xneo4j
+ xredis

### Dependencies

#### xgorm

+ See [go.mod](./xgorm/go.mod) and [go.sum](./xgorm/go.sum)
+ `github.com/Aoi-hosizora/ahlib v1.6.0`
+ `github.com/Aoi-hosizora/ahlib-db/xdbutils v1.6.0`
+ `github.com/jinzhu/gorm v1.9.16`
+ `github.com/go-sql-driver/mysql v1.5.0`
+ `github.com/VividCortex/mysqlerr v1.0.0`
+ `github.com/mattn/go-sqlite3 v1.14.0`
+ `github.com/lib/pq v1.1.1`
+ `github.com/sirupsen/logrus v1.8.1`

#### xgormv2

+ See [go.mod](./xgormv2/go.mod) and [go.sum](./xgormv2/go.sum)
+ `github.com/Aoi-hosizora/ahlib v1.6.0`
+ `github.com/Aoi-hosizora/ahlib-db/xdbutils v1.6.0`
+ `gorm.io/gorm v1.22.4`
+ `gorm.io/driver/mysql v1.2.3`
+ `gorm.io/driver/sqlite v1.2.6`
+ `github.com/go-sql-driver/mysql v1.6.0`
+ `github.com/VividCortex/mysqlerr v1.0.0`
+ `github.com/mattn/go-sqlite3 v1.14.9`
+ `github.com/sirupsen/logrus v1.8.1`

#### xneo4j

+ See [go.mod](./xneo4j/go.mod) and [go.sum](./xneo4j/go.sum)
+ `github.com/Aoi-hosizora/ahlib v1.6.0`
+ `github.com/Aoi-hosizora/ahlib-db/xdbutils v1.6.0`
+ `github.com/neo4j/neo4j-go-driver v1.8.3`
+ `github.com/sirupsen/logrus v1.8.1`

#### xredis

+ See [go.mod](./xredis/go.mod) and [go.sum](./xredis/go.sum)
+ `github.com/Aoi-hosizora/ahlib v1.6.0`
+ `github.com/go-redis/redis/v8 v8.4.11`
+ `github.com/sirupsen/logrus v1.8.1`
