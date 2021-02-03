# xgorm

## Dependencies

+ github.com/Aoi-hosizora/ahlib
+ github.com/jinzhu/gorm

## Documents

### Types

+ `type GormTime struct`
+ `type GormTime2 struct`
+ `type PropertyValue struct`
+ `type PropertyDict map`

### Variables

+ None

### Constants

+ `const DefaultDeletedAtTimestamp string`
+ `const MySQLDuplicateEntryErrno int`
+ `const SQLiteUniqueConstraintErrno int`
+ `const PostgreSQLUniqueViolationErrno string`

### Functions

+ `func HookDeletedAt(db *gorm.DB, deletedAtTimestamp string)`
+ `func IsMySQL(db *gorm.DB) bool`
+ `func IsSQLite(db *gorm.DB) bool`
+ `func IsPostgreSQL(db *gorm.DB) bool`
+ `func IsMySQLDuplicateEntryError(err error) bool`
+ `func IsSQLiteUniqueConstraintError(err error) bool`
+ `func IsPostgreSQLUniqueViolationError(err error) bool`
+ `func QueryErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue`
+ `func GenerateOrderByExp(source string, dict PropertyDict) string`
+ `func GenerateOrderByFunc(dict PropertyDict) func(source string) string`

### Methods

+ `func (p *PropertyValue) Destinations() []string`
+ `func (p *PropertyValue) Reverse() bool`

---

+ `type SilenceLogger struct{}`
+ `NewSilenceLogger() *GormSilenceLogger`
+ `type LogrusLogger struct {}`
+ `NewLogrusLogger(logger *logrus.Logger) *GormLogrus`
+ `type StdLogLogger struct {}`
+ `NewStdLogLogger(logger *log.Logger) *GormLogger`
