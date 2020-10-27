package xgorm

import (
	"github.com/Aoi-hosizora/ahlib/xproperty"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
)

type TblTest struct {
	Id   uint64
	Name string
}

type TblTestUnique struct {
	Id   uint64
	Name string `gorm:"unique_index:nk_name"`
}

func TestLogrus(t *testing.T) {
	db, err := gorm.Open("mysql", "root:123@tcp(localhost:3306)/db_test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatalln(err)
	}

	db.SingularTable(true)
	db.LogMode(true)

	db.SetLogger(NewGormLogrus(logrus.New()))
	HookDeleteAtField(db, DefaultDeleteAtTimestamp)

	test := &TblTest{}
	db.Model(&TblTest{}).First(test)
	log.Println(test)
	tests := make([]*TblTest, 0)
	db.Model(&TblTest{}).Find(&tests)
	log.Println(tests)
	db.Model(test).Related(test)
	log.Println(tests)
}

func TestLogger(t *testing.T) {
	db, err := gorm.Open("mysql", "root:123@tcp(localhost:3306)/db_test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatalln(err)
	}

	db.SingularTable(true)
	db.LogMode(true)

	db.SetLogger(NewGormSilenceLogger())
	db.SetLogger(NewGormLogger(log.New(os.Stderr, "", log.LstdFlags)))
	HookDeleteAtField(db, DefaultDeleteAtTimestamp)

	test := &TblTest{}
	db.Model(&TblTest{}).First(test)
	log.Println(test)
	tests := make([]*TblTest, 0)
	db.Model(&TblTest{}).Find(&tests)
	log.Println(tests)
}

func TestOthers(t *testing.T) {
	db, err := gorm.Open("mysql", "root:123@tcp(localhost:3306)/db_test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatalln(err)
	}

	db.SingularTable(true)
	db.LogMode(false)

	db.AutoMigrate(&TblTestUnique{})

	rdb := db.Create(&TblTestUnique{Id: 1, Name: "1"})
	log.Println(rdb.Error)
	rdb = db.Create(&TblTestUnique{Id: 2, Name: "2"})
	log.Println(rdb.Error)
	rdb = db.Create(&TblTestUnique{Id: 3, Name: "1"})
	log.Println(rdb.Error)
	log.Println(IsMySQLDuplicateEntryError(rdb.Error))

	tt := &TblTestUnique{}
	rdb = db.Where("id = ?", 1).First(tt)
	log.Println(QueryErr(rdb))
	rdb = db.Where("id = ?", 3).First(tt)
	log.Println(QueryErr(rdb))
}

func TestOrderBy(t *testing.T) {
	dict := map[string]*xproperty.PropertyMapperValue{
		"uid":      xproperty.NewValue(false, "uid"),
		"username": xproperty.NewValue(false, "lastName", "firstName"),
		"age":      xproperty.NewValue(true, "birthday"),
	}

	xtesting.Equal(t, OrderByFunc(dict)("uid desc,age,username"), "uid DESC, birthday DESC, lastName ASC, firstName ASC")
	xtesting.Equal(t, OrderByFunc(dict)(""), "")
	xtesting.Equal(t, OrderByFunc(dict)("a"), "")
	xtesting.Equal(t, OrderByFunc(map[string]*xproperty.PropertyMapperValue{})("a"), "")
}
