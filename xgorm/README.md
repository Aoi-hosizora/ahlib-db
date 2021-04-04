# xgorm

## Dependencies

+ github.com/Aoi-hosizora/ahlib
+ github.com/jinzhu/gorm
+ github.com/go-sql-driver/mysql
+ github.com/mattn/go-sqlite3 (cgo)
+ github.com/lib/pq
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type GormTime struct`
+ `type GormTime2 struct`
+ `type PropertyValue struct`
+ `type PropertyDict map`
+ `type ILogger interface`
+ `type LoggerOption func`
+ `type SilenceLogger struct`
+ `type LogrusLogger struct`
+ `type LoggerLogger struct`

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
+ `func IsSQLiteUniqueConstraintError(err error) bool // cgo`
+ `func IsPostgreSQLUniqueViolationError(err error) bool`
+ `func QueryErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error) // !cgo+cgo`
+ `func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error) // !cgo+cgo`
+ `func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue`
+ `func GenerateOrderByExp(source string, dict PropertyDict) string`
+ `func WithLogInfo(logInfo bool) LoggerOption`
+ `func WithLogOther(logOther bool) LoggerOption`
+ `func NewSilenceLogger() *SilenceLogger`
+ `func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger`
+ `func NewLoggerLogger(logger logrus.StdLogger, options ...LoggerOption) *LoggerLogger`

### Methods

+ `func (p *PropertyValue) Destinations() []string`
+ `func (p *PropertyValue) Reverse() bool`
+ `func (g *SilenceLogger) Print(...interface{})`
+ `func (g *LogrusLogger) Print(v ...interface{})`
+ `func (g *LoggerLogger) Print(v ...interface{})`
