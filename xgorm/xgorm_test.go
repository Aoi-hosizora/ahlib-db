package xgorm

import (
	"errors"
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
	"time"
)

const (
	mysqlDsl   = "root:123@tcp(localhost:3306)/db_test?charset=utf8&parseTime=True&loc=Local"
	sqliteFile = "test.sql"
)

type User struct {
	Uid  int    `gorm:"primary_key; auto_increment"`
	Name string `gorm:"not null; unique_index:uk_name"`
	GormTime
}

func TestHook(t *testing.T) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	check := func(db *gorm.DB, write bool) error {
		if db.Error != nil {
			return db.Error
		}
		if !write && db.RecordNotFound() {
			return errors.New("record not found")
		}
		if write && db.RowsAffected == 0 {
			return errors.New("rows affected is zero")
		}
		return nil
	}

	for _, tc := range []struct {
		giveDialect string
		giveParam   string
	}{
		{"mysql", mysqlDsl},
		{"sqlite3", sqliteFile},
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			db, err := gorm.Open(tc.giveDialect, tc.giveParam)
			if err != nil {
				log.Println(err)
				t.FailNow()
			}
			db.LogMode(true)
			db.SetLogger(NewLogrusLogger(l))
			HookDeletedAt(db, DefaultDeletedAtTimestamp)
			db.DropTableIfExists(&User{})
			if db.AutoMigrate(&User{}).Error != nil {
				log.Println(err)
				t.FailNow()
			}

			// create
			user := &User{Uid: 1, Name: "user1"}
			db.Model(&User{}).Create(user)
			xtesting.Equal(t, user.DeletedAt.Format("2006-01-02 15:04:05"), DefaultDeletedAtTimestamp)

			// query
			user = &User{}
			xtesting.Nil(t, check(db.Model(&User{}).Where(&User{Uid: 1}).First(user), false))
			xtesting.Equal(t, user.Uid, 1)
			xtesting.Equal(t, user.Name, "user1")
			xtesting.Equal(t, user.DeletedAt.Format("2006-01-02 15:04:05"), DefaultDeletedAtTimestamp)

			// update
			xtesting.Nil(t, check(db.Model(&User{Uid: 1}).Updates(&User{Name: "user1_new"}), true))
			user = &User{}
			xtesting.Nil(t, check(db.Model(&User{}).Where(&User{Uid: 1}).First(user), false))
			xtesting.Equal(t, user.Uid, 1)
			xtesting.Equal(t, user.Name, "user1_new")

			// soft delete
			xtesting.Nil(t, check(db.Model(&User{}).Delete(&User{Uid: 1}), true))
			user = &User{}
			xtesting.NotNil(t, check(db.Model(&User{}).Where(&User{Uid: 1}).First(user), false))
			xtesting.Nil(t, check(db.Unscoped().Model(&User{}).Where(&User{Uid: 1}).First(user), false))
			xtesting.NotEqual(t, user.DeletedAt.Format("2006-01-02 15:04:05"), DefaultDeletedAtTimestamp)

			// hard delete
			xtesting.Nil(t, check(db.Unscoped().Model(&User{}).Delete(&User{Uid: 1}), true))
			xtesting.NotNil(t, check(db.Unscoped().Model(&User{}).Where(&User{Uid: 1}).First(&User{}), false))
		})
	}
}

func TestHelper(t *testing.T) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})

	for _, tc := range []struct {
		giveDialect string
		giveParam   string
	}{
		{"mysql", mysqlDsl},
		{"sqlite3", sqliteFile},
	} {
		t.Run(tc.giveDialect, func(t *testing.T) {
			db, err := gorm.Open(tc.giveDialect, tc.giveParam)
			if err != nil {
				log.Println(err)
				t.FailNow()
			}
			db.LogMode(true)
			db.SetLogger(NewLogrusLogger(l))
			HookDeletedAt(db, DefaultDeletedAtTimestamp)
			db.DropTableIfExists(&User{})
			if db.AutoMigrate(&User{}).Error != nil {
				log.Println(err)
				t.FailNow()
			}

			// create
			sts, err := CreateErr(db.Model(&User{}).Create(&User{Uid: 1, Name: "user1"}))
			xtesting.Equal(t, sts, xstatus.DbSuccess)
			xtesting.Nil(t, err)
			sts, err = CreateErr(db.Model(&User{}).Create(&User{Uid: 2, Name: "user1"})) // existed
			xtesting.Equal(t, sts, xstatus.DbExisted)
			xtesting.NotEqual(t, err, nil)
			sts, err = CreateErr(db.Model(&User{}).Create(&User{Uid: 2, Name: "user2"}))
			xtesting.Equal(t, sts, xstatus.DbSuccess)
			xtesting.Nil(t, err)

			// query
			sts, err = QueryErr(db.Model(&User{}).Where(&User{Uid: 1}).First(&User{}))
			xtesting.Equal(t, sts, xstatus.DbSuccess)
			xtesting.Nil(t, err)
			sts, err = QueryErr(db.Model(&User{}).Where(&User{Uid: 2, Name: "user1"}).First(&User{})) // not found
			xtesting.Equal(t, sts, xstatus.DbNotFound)
			xtesting.Nil(t, err)
			sts, err = QueryErr(db.Model(&User{}).Where(&User{Uid: 3}).First(&User{})) // not found
			xtesting.Equal(t, sts, xstatus.DbNotFound)
			xtesting.Nil(t, err)

			// update
			sts, err = UpdateErr(db.Model(&User{}).Where(&User{Uid: 1}).Updates(&User{Name: "user1_new"}))
			xtesting.Equal(t, sts, xstatus.DbSuccess)
			xtesting.Nil(t, err)
			sts, err = UpdateErr(db.Model(&User{}).Where(&User{Uid: 3}).Updates(&User{Name: "user3"})) // not found
			xtesting.Equal(t, sts, xstatus.DbNotFound)
			xtesting.Nil(t, err)
			sts, err = UpdateErr(db.Model(&User{}).Where(&User{Uid: 2}).Updates(&User{Name: "user1_new"})) // existed
			xtesting.Equal(t, sts, xstatus.DbExisted)
			xtesting.NotNil(t, err)

			// delete
			sts, err = DeleteErr(db.Model(&User{}).Delete(&User{Uid: 1}))
			xtesting.Equal(t, sts, xstatus.DbSuccess)
			xtesting.Nil(t, err)
			sts, err = DeleteErr(db.Model(&User{}).Delete(&User{Uid: 2}))
			xtesting.Equal(t, sts, xstatus.DbSuccess)
			xtesting.Nil(t, err)
			sts, err = DeleteErr(db.Model(&User{}).Delete(&User{Uid: 3})) // not found
			xtesting.Equal(t, sts, xstatus.DbNotFound)
			xtesting.Nil(t, err)

			// order
			dict := PropertyDict{
				"uid":      NewPropertyValue(false, "uid"),
				"username": NewPropertyValue(false, "firstname", "lastname"),
				"age":      NewPropertyValue(true, "birthday"),
			}
			for _, tc := range []struct {
				giveSource string
				giveDict   PropertyDict
				want       string
			}{
				{"uid, xxx", dict, "uid ASC"},
				{"uid desc xxx", dict, "uid DESC"},
				{"uid, username", dict, "uid ASC, firstname ASC, lastname ASC"},
				{"username desc, age desc", dict, "firstname DESC, lastname DESC, birthday ASC"},
			} {
				xtesting.Equal(t, GenerateOrderByExp(tc.giveSource, tc.giveDict), tc.want)
			}
		})
	}
}

func TestLogger(t *testing.T) {
	l1 := logrus.New()
	l1.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	l2 := log.New(os.Stderr, "", log.LstdFlags)

	for _, tc := range []struct {
		name   string
		mode   bool
		logger ILogger
	}{
		{"disable mode", false, nil},
		{"default", true, nil},
		{"silence", true, NewSilenceLogger()},
		{"logrus", true, NewLogrusLogger(l1)},
		{"logger", true, NewLoggerLogger(l2)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open("mysql", mysqlDsl)
			if err != nil {
				log.Println(err)
				t.FailNow()
			}
			db.LogMode(tc.mode)
			if tc.logger != nil {
				db.SetLogger(tc.logger)
			}
			HookDeletedAt(db, DefaultDeletedAtTimestamp) // log [info]
			db.DropTableIfExists(&User{})
			if db.AutoMigrate(&User{}).Error != nil {
				log.Println(err)
				t.FailNow()
			}

			db.Create(&User{Uid: 1, Name: "user1"})
			db.Create(&User{Uid: 1, Name: "user1"}) // log [log]
			db.Model(&User{}).Where(&User{Uid: 1}).First(&User{})
			db.Model(&User{}).Where("name = ? OR name = ?", []byte("user1"), []byte{0x00, 0x01}).First(&User{}) // ?
			db.Model(&User{}).Where("deleted_at = $1 OR deleted_at = $2", time.Time{}, nil).First(&User{})      // $
		})
	}
}
