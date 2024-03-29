# ahlib-db

[![Build Status](https://travis-ci.com/Aoi-hosizora/ahlib-db.svg?branch=master)](https://travis-ci.com/Aoi-hosizora/ahlib-db)
[![codecov](https://codecov.io/gh/Aoi-hosizora/ahlib-db/branch/master/graph/badge.svg)](https://codecov.io/gh/Aoi-hosizora/ahlib-db)
[![Go Report Card](https://goreportcard.com/badge/github.com/Aoi-hosizora/ahlib-db)](https://goreportcard.com/report/github.com/Aoi-hosizora/ahlib-db)
[![License](http://img.shields.io/badge/license-mit-blue.svg)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/Aoi-hosizora/ahlib-db)](https://github.com/Aoi-hosizora/ahlib-db/releases)

+ ATTENTION: This package is **ARCHIVED**, please refer to https://github.com/Aoi-hosizora/ahlib-mx and use `ahlib-mx` (origin: `ahlib-web`) instead.
+ A personal golang library for db development, including gorm (mysql, sqlite, postgresql), redis (go-redis), neo4j (neo4j-go-driver).

### Related libraries

+ [Aoi-hosizora/ahlib](https://github.com/Aoi-hosizora/ahlib)
+ [Aoi-hosizora/ahlib-more](https://github.com/Aoi-hosizora/ahlib-more)
+ [Aoi-hosizora/ahlib-web](https://github.com/Aoi-hosizora/ahlib-web)
+ [Aoi-hosizora/ahlib-db](https://github.com/Aoi-hosizora/ahlib-db)

### Packages

+ xgorm
+ xneo4j
+ xredis

### Dependencies

+ See [go.mod](./go.mod) and [go.sum](./go.sum)
+ `github.com/Aoi-hosizora/ahlib v1.5.0`
+ `github.com/jinzhu/gorm v1.9.15`
+ `github.com/go-sql-driver/mysql v1.5.0`
+ `github.com/mattn/go-sqlite3 v1.14.0`
+ `github.com/lib/pq v1.1.1`
+ `github.com/go-redis/redis/v8 v8.4.11`
+ `github.com/neo4j/neo4j-go-driver v1.8.3`
+ `github.com/sirupsen/logrus v1.7.0`
