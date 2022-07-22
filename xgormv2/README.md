# xgormv2

## Dependencies

+ github.com/Aoi-hosizora/ahlib
+ gorm.io/gorm
+ gorm.io/driver/mysql
+ gorm.io/driver/sqlite (cgo)
+ github.com/go-sql-driver/mysql
+ github.com/VividCortex/mysqlerr
+ github.com/mattn/go-sqlite3 (cgo)
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type MySQLConfig struct`
+ `type PropertyValue struct`
+ `type PropertyDict map`
+ `type GormTime struct`
+ `type GormTime2 struct`
+ `type LoggerOption func`
+ `type ILogger interface`
+ `type SilenceLogger struct`
+ `type LogrusLogger struct`
+ `type StdLogger struct`
+ `type LoggerParam struct`

### Variables

+ `var IsPostgreSQLUniqueViolationError func`
+ `var FormatLoggerFunc func`
+ `var FieldifyLoggerFunc func`

### Constants

+ `const MySQL string`
+ `const SQLite string`
+ `const Postgres string`
+ `const MySQLDuplicateEntryErrno int`
+ `const SQLiteUniqueConstraintErrno int`
+ `const PostgreSQLUniqueViolationErrno string`
+ `const DefaultDeletedAtTimestamp string`
+ `const CreateCallbackName string`
+ `const UpdateCallbackName string`
+ `const DeleteCallbackName string`
+ `const QueryCallbackName string`
+ `const RowQueryCallbackName string`

### Functions

+ `func IsMySQL(db *gorm.DB) bool`
+ `func IsSQLite(db *gorm.DB) bool`
+ `func IsPostgreSQL(db *gorm.DB) bool`
+ `func MySQLDefaultCharsetTimeLocParam() map[string]string`
+ `func MySQLDefaultDsn(username, password, address, database string) string`
+ `func SQLiteDefaultDsn(filename string) string`
+ `func PostgresDefaultDsn(username, password, host string, port int, database string) string`
+ `func IsMySQLDuplicateEntryError(err error) bool`
+ `func IsSQLiteUniqueConstraintError(err error) bool` (cgo)
+ `func IsRecordNotFound(err error) bool`
+ `func QueryErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue`
+ `func GenerateOrderByExp(source string, dict PropertyDict) string`
+ `func HookDeletedAt(db *gorm.DB, defaultTimestamp string)`
+ `func WithLogInfo(log bool) LoggerOption`
+ `func WithLogSQL(log bool) LoggerOption`
+ `func WithLogOther(log bool) LoggerOption`
+ `func WithSlowThreshold(threshold time.Duration) LoggerOption`
+ `func EnableLogger()`
+ `func DisableLogger()`
+ `func NewSilenceLogger() *SilenceLogger`
+ `func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger`
+ `func NewStdLogger(logger logrus.StdLogger, options ...LoggerOption) *StdLogger`

### Methods

+ `func (p *PropertyValue) Destinations() []string`
+ `func (p *PropertyValue) Reverse() bool`
+ `func (s *SilenceLogger) Print(...interface{})`
+ `func (l *LogrusLogger) Print(v ...interface{})`
+ `func (l *StdLogger) Print(v ...interface{})`