# xgorm

### Functions

#### Normal

+ `const DefaultDeletedAtTimestamp string`
+ `HookDeletedAtField(db *gorm.DB, deletedAtTimestamp string)`
+ `type GormTime struct {}`
+ `type GormTime2 struct {}`
+ `type SilenceLogger struct{}`
+ `NewSilenceLogger() *GormSilenceLogger`
+ `type LogrusLogger struct {}`
+ `NewLogrusLogger(logger *logrus.Logger) *GormLogrus`
+ `type StdLogLogger struct {}`
+ `NewStdLogLogger(logger *log.Logger) *GormLogger`

#### Helper

+ `type Helper struct {}`
+ `WithDB(db *gorm.DB) *Helper`
+ `(h *Helper) GetDB() *gorm.DB`
+ `(h *Helper) Pagination(limit int32, page int32) *gorm.DB`
+ `(h *Helper) Count(model interface{}, where interface{}) (uint64, error)`
+ `(h *Helper) Exist(model interface{}, where interface{}) (bool, error)`
+ `IsMySQL(db *gorm.DB) bool`
+ `const MySQLDuplicateEntryError int`
+ `IsMySQLDuplicateEntryError(err error) bool`
+ `QueryErr(rdb *gorm.DB) (bool, error)`
+ `CreateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `UpdateErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `DeleteErr(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `OrderByFunc(p xproperty.PropertyDict) func(source string) string`
