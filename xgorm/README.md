# xgorm

### Functions

#### Normal

+ `const DefaultDeleteAtTimestamp string`
+ `HookDeleteAtField(db *gorm.DB, defaultDeleteAtTimestamp string)`
+ `type GormTime struct {}`
+ `type GormTimeWithoutDeleteAt struct {}`
+ `IsMySQLDuplicateEntryError(err error) bool`
+ `type GormSilenceLogger struct{}`
+ `NewGormSilenceLogger() *GormSilenceLogger`
+ `type GormLogrus struct {}`
+ `NewGormLogrus(logger *logrus.Logger) *GormLogrus`
+ `type GormLogger struct {}`
+ `NewGormLogger(logger *log.Logger) *GormLogger`

#### Helper

+ `type Helper struct {}`
+ `WithDB(db *gorm.DB) *Helper`
+ `(h *Helper) Pagination(limit int32, page int32) *gorm.DB`
+ `(h *Helper) Count(model interface{}, where interface{}) (uint64, error)`
+ `(h *Helper) Exist(model interface{}, where interface{}) (bool, error)`
+ `(h *Helper) Create(model interface{}, object interface{}) (xstatus.DbStatus, error)`
+ `(h *Helper) Update(model interface{}, where interface{}, object interface{}) (xstatus.DbStatus, error)`
+ `(h *Helper) Delete(model interface{}, where interface{}, object interface{}) (xstatus.DbStatus, error)`
+ `IsMySQL(db *gorm.DB) bool`
+ `QueryErr(rdb *gorm.DB) (bool, error)`
+ `CreateDB(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `UpdateDB(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `DeleteDB(rdb *gorm.DB) (xstatus.DbStatus, error)`
+ `OrderByFunc(p xproperty.PropertyDict) func(source string) string`
