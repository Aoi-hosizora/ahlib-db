# xgorm

## Dependencies

+ github.com/Aoi-hosizora/ahlib
+ github.com/jinzhu/gorm
+ github.com/go-sql-driver/mysql
+ github.com/VividCortex/mysqlerr
+ github.com/mattn/go-sqlite3 (cgo)
+ github.com/lib/pq
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type GormTime struct`
+ `type GormTime2 struct`
+ `type MySQLConfig struct`
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
+ `const MySQL string`
+ `const SQLite string`
+ `const Postgres string`
+ `const MySQLDuplicateEntryErrno int`
+ `const SQLiteUniqueConstraintErrno int`
+ `const PostgreSQLUniqueViolationErrno string`

### Functions

+ `func HookDeletedAt(db *gorm.DB, deletedAtTimestamp string) *gorm.DB`
+ `func IsMySQL(db *gorm.DB) bool`
+ `func IsSQLite(db *gorm.DB) bool`
+ `func IsPostgreSQL(db *gorm.DB) bool`
+ `func MySQLDefaultDsn(username, password, address, database string) string`
+ `func SQLiteDefaultDsn(filename string) string`
+ `func PostgresDefaultDsn(username, password, host string, port int, database string) string`
+ `func IsMySQLDuplicateEntryError(err error) bool`
+ `func IsSQLiteUniqueConstraintError(err error) bool // cgo`
+ `func IsPostgreSQLUniqueViolationError(err error) bool`
+ `func QueryErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue`
+ `func GenerateOrderByExp(source string, dict PropertyDict) string`
+ `func WithLogInfo(logInfo bool) LoggerOption`
+ `func WithLogSql(logSql bool) LoggerOption`
+ `func WithLogOther(logOther bool) LoggerOption`
+ `func EnableLogger()`
+ `func DisableLogger()`
+ `func NewSilenceLogger() *SilenceLogger`
+ `func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger`
+ `func NewLoggerLogger(logger logrus.StdLogger, options ...LoggerOption) *LoggerLogger`

### Methods

+ `func (p *PropertyValue) Destinations() []string`
+ `func (p *PropertyValue) Reverse() bool`
+ `func (g *SilenceLogger) Print(...interface{})`
+ `func (g *LogrusLogger) Print(v ...interface{})`
+ `func (g *LoggerLogger) Print(v ...interface{})`
