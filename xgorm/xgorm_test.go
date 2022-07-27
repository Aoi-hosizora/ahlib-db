package xgorm

import (
	"errors"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"log"
	"testing"
	"time"
)

func TestMass1(t *testing.T) {
	xtesting.Equal(t, MySQLDefaultDsn("root", "123", "localhost:3306", "db_test"),
		"root:123@tcp(localhost:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local")
	xtesting.Equal(t, SQLiteDefaultDsn("test.sql"), "test.sql")
	xtesting.Equal(t, PostgreSQLDefaultDsn("postgres", "123", "localhost", 5432, "db_test"),
		"host=localhost port=5432 user=postgres password=123 dbname=db_test")

	xtesting.True(t, IsMySQLDuplicateEntryError(&mysql.MySQLError{Number: MySQLDuplicateEntryErrno}))
	xtesting.True(t, IsPostgreSQLUniqueViolationError(pq.Error{Code: PostgreSQLUniqueViolationErrno}))
	xtesting.True(t, IsPostgreSQLUniqueViolationError(&pq.Error{Code: PostgreSQLUniqueViolationErrno}))
}

var (
	mysqlDsn   = MySQLDefaultDsn("root", "123", "localhost:3306", "db_test")
	sqliteFile = SQLiteDefaultDsn("test.sql")
)

type User struct {
	Uid  int    `gorm:"primary_key; auto_increment"`
	Name string `gorm:"type:varchar(12); not null; unique_index:uk_name"`
	GormTime
}

func testHook(t *testing.T, giveDialect, giveParam string) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	check := func(db *gorm.DB, write bool) error {
		switch {
		case db.Error != nil:
			return db.Error
		case !write && db.RecordNotFound():
			return errors.New("record not found")
		case write && db.RowsAffected == 0:
			return errors.New("rows affected is zero")
		}
		return nil
	}

	db, err := gorm.Open(giveDialect, giveParam)
	if !xtesting.Nil(t, err) {
		t.FailNow()
	}
	_ = CreateCallbackName
	db.LogMode(true)
	db.SetLogger(NewLogrusLogger(l))
	HookDeletedAt(db, DefaultDeletedAtTimestamp)
	db.DropTableIfExists(&User{})
	defer db.DropTableIfExists(&User{})
	if !xtesting.Nil(t, db.AutoMigrate(&User{}).Error) {
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
}

func testHelper(t *testing.T, giveDialect, giveParam string) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})

	db, err := gorm.Open(giveDialect, giveParam)
	if !xtesting.Nil(t, err) {
		t.FailNow()
	}
	db.LogMode(true)
	db.SetLogger(NewLogrusLogger(l))
	HookDeletedAt(db, DefaultDeletedAtTimestamp)
	db.DropTableIfExists(&User{})
	defer db.DropTableIfExists(&User{})
	if !xtesting.Nil(t, db.AutoMigrate(&User{}).Error) {
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

	// hack
	db.Error = errors.New("test")
	sts, err = QueryErr(db)
	xtesting.Equal(t, sts, xstatus.DbFailed)
	xtesting.Equal(t, err.Error(), "test")
	sts, err = CreateErr(db)
	xtesting.Equal(t, sts, xstatus.DbFailed)
	xtesting.Equal(t, err.Error(), "test")
	sts, err = UpdateErr(db)
	xtesting.Equal(t, sts, xstatus.DbFailed)
	xtesting.Equal(t, err.Error(), "test")
	sts, err = DeleteErr(db)
	xtesting.Equal(t, sts, xstatus.DbFailed)
	xtesting.Equal(t, err.Error(), "test")
	db.Error = nil

	// order
	dict := PropertyDict{
		"uid":      NewPropertyValue(false, "uid"),
		"username": NewPropertyValue(false, "firstname", "lastname"),
		"age":      NewPropertyValue(true, "birthday"),
	}
	nilOptions := []OrderByOption{
		WithOrderBySourceSeparator(""), WithOrderByTargetSeparator(""),
		WithOrderBySourceProcessor(nil), WithOrderByTargetProcessor(nil),
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
		xtesting.Equal(t, GenerateOrderByExpr(tc.giveSource, tc.giveDict, nilOptions...), tc.want)
	}
}

func testLogger(t *testing.T, giveDialect, giveParam string) {
	l1 := logrus.New()
	l1.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	l2 := log.Default()

	for _, tc := range []struct {
		name   string
		mode   bool
		enable bool
		custom bool
		logger ILogger
	}{
		// {"default", true, true, false, nil},
		{"disable_mode", false, true, false, nil},
		{"silence", true, true, false, NewSilenceLogger()},
		//
		{"logrus", true, true, false, NewLogrusLogger(l1, WithSlowThreshold(time.Millisecond*10))},
		{"logrus_custom", true, true, true, NewLogrusLogger(l1)},
		{"logrus_no_info_other", true, true, false, NewLogrusLogger(l1, WithLogInfo(false), WithLogOther(false))},
		{"logrus_no_sql", true, true, false, NewLogrusLogger(l1, WithLogSQL(false))},
		{"logrus_disable", true, false, false, NewLogrusLogger(l1)},
		//
		{"stdlog", true, true, false, NewStdLogger(l2, WithSlowThreshold(time.Millisecond*10))},
		{"stdlog_custom", true, true, true, NewStdLogger(l2)},
		{"stdlog_no_xxx", true, true, false, NewStdLogger(l2, WithLogInfo(false), WithLogSQL(false), WithLogOther(false))},
		{"stdlog_disable", true, false, false, NewStdLogger(l2)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(giveDialect, giveParam)
			if !xtesting.Nil(t, err) {
				t.FailNow()
			}
			db.LogMode(tc.mode)
			if tc.enable {
				EnableLogger()
			} else {
				DisableLogger()
			}
			if tc.logger != nil {
				db.SetLogger(tc.logger)
			}
			if tc.custom {
				FormatLoggerFunc = func(p *LoggerParam) string {
					if p.Type != "sql" {
						return fmt.Sprintf("[Gorm] msg: %s", p.Message)
					}
					return fmt.Sprintf("[Gorm] %7d - %12s - %s - %s", p.Rows, p.Duration.String(), p.SQL, p.Source)
				}
				FieldifyLoggerFunc = func(p *LoggerParam) logrus.Fields {
					return logrus.Fields{"module": "gorm", "type": p.Type}
				}
				defer func() {
					FormatLoggerFunc = nil
					FieldifyLoggerFunc = nil
				}()
			}

			HookDeletedAt(db, DefaultDeletedAtTimestamp) // log [info]
			db.DropTableIfExists(&User{})
			defer db.DropTableIfExists(&User{})
			if !xtesting.Nil(t, db.AutoMigrate(&User{}).Error) {
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
