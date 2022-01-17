package xgorm

import (
	"errors"
	"github.com/Aoi-hosizora/ahlib/xstatus"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"testing"
	"time"
)

func TestMess1(t *testing.T) {
	xtesting.Equal(t, MySQLDefaultDsn("root", "123", "localhost:3306", "db_test"),
		"root:123@tcp(localhost:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local")
	xtesting.Equal(t, SQLiteDefaultDsn("test.sql"), "test.sql")
	xtesting.Equal(t, PostgresDefaultDsn("postgres", "123", "localhost", 5432, "db_test"),
		"host=localhost port=5432 user=postgres password=123 dbname=db_test")

	xtesting.True(t, IsMySQLDuplicateEntryError(&mysql.MySQLError{Number: MySQLDuplicateEntryErrno}))
	IsPostgreSQLUniqueViolationError = func(err error) bool {
		return err.Error() == PostgreSQLUniqueViolationErrno
	}
	defer func() { IsPostgreSQLUniqueViolationError = nil }()
	xtesting.True(t, IsPostgreSQLUniqueViolationError(errors.New("23505")))
}

var (
	mysqlDsn   = MySQLDefaultDsn("root", "123", "localhost:3306", "db_test")
	sqliteFile = SQLiteDefaultDsn("test.sql")
)

type User struct {
	Uid  int    `gorm:"primaryKey; autoIncrement"`
	Name string `gorm:"type:varchar(12); not null; uniqueIndex:uk_name"`
	GormTime
}

func testHook(t *testing.T, giveDialector gorm.Dialector) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	check := func(db *gorm.DB, write bool) error {
		switch {
		case db.Error != nil:
			return db.Error
		case !write && IsRecordNotFound(db.Error):
			return errors.New("record not found")
		case write && db.RowsAffected == 0:
			return errors.New("rows affected is zero")
		}
		return nil
	}

	db, err := gorm.Open(giveDialector, &gorm.Config{
		Logger: logger.New(log.Default(), logger.Config{
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		}),
	})
	if !xtesting.Nil(t, err) {
		t.FailNow()
	}
	_ = CreateCallbackName
	// db.LogMode(true)
	// db.SetLogger(NewLogrusLogger(l))
	HookDeletedAt(db, DefaultDeletedAtTimestamp)
	db.Migrator().DropTable(&User{})
	defer db.Migrator().DropTable(&User{})
	if !xtesting.Nil(t, db.AutoMigrate(&User{})) {
		t.FailNow()
	}

	// create
	user := &User{Uid: 1, Name: "user1"}
	db.Model(&User{}).Create(user)
	if user.DeletedAt.Valid { // <<<
		xtesting.Equal(t, user.DeletedAt.Time.Format("2006-01-02 15:04:05"), DefaultDeletedAtTimestamp)
	}

	// query
	user = &User{}
	xtesting.Nil(t, check(db.Model(&User{}).Where(&User{Uid: 1}).First(user), false))
	xtesting.Equal(t, user.Uid, 1)
	xtesting.Equal(t, user.Name, "user1")
	xtesting.Equal(t, user.DeletedAt.Time.Format("2006-01-02 15:04:05"), DefaultDeletedAtTimestamp)

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
	xtesting.NotEqual(t, user.DeletedAt.Time.Format("2006-01-02 15:04:05"), DefaultDeletedAtTimestamp)

	// hard delete
	xtesting.Nil(t, check(db.Unscoped().Model(&User{}).Delete(&User{Uid: 1}), true))
	xtesting.NotNil(t, check(db.Unscoped().Model(&User{}).Where(&User{Uid: 1}).First(&User{}), false))
}

func testHelper(t *testing.T, giveDialector gorm.Dialector) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})

	db, err := gorm.Open(giveDialector, &gorm.Config{
		Logger: logger.New(log.Default(), logger.Config{
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		}),
	})
	if !xtesting.Nil(t, err) {
		t.FailNow()
	}
	// db.LogMode(true)
	// db.SetLogger(NewLogrusLogger(l))
	HookDeletedAt(db, DefaultDeletedAtTimestamp)
	db.Migrator().DropTable(&User{})
	defer db.Migrator().DropTable(&User{})
	if !xtesting.Nil(t, db.AutoMigrate(&User{})) {
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
}
