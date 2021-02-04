# xneo4j

## Dependencies

+ github.com/neo4j/neo4j-go-driver

## Documents

### Types

+ `type PropertyValue struct`
+ `type PropertyDict map`
+ `type DialHandler func`
+ `type Pool struct`

### Variables

+ None

### Constants

+ None

### Functions

+ `func GetByColumnIndex(records []neo4j.Record, row int, index int) (interface{}, bool)`
+ `func GetByColumnKey(records []neo4j.Record, row int, key string) (interface{}, bool)`
+ `func GetInteger(data interface{}) int64`
+ `func GetFloat(data interface{}) float64`
+ `func GetString(data interface{}) string`
+ `func GetBoolean(data interface{}) bool`
+ `func GetByteArray(data interface{}) []byte`
+ `func GetList(data interface{}) []interface{}`
+ `func GetMap(data interface{}) map[string]interface{}`
+ `func GetNode(data interface{}) neo4j.Node`
+ `func GetRel(data interface{}) neo4j.Relationship`
+ `func GetPath(data interface{}) neo4j.Path`
+ `func GetPoint(data interface{}) neo4j.Point`
+ `func GetDate(data interface{}) neo4j.Date`
+ `func GetTime(data interface{}) neo4j.OffsetTime`
+ `func GetLocalTime(data interface{}) neo4j.LocalTime`
+ `func GetDateTime(data interface{}) time.Time`
+ `func GetLocalDateTime(data interface{}) neo4j.LocalDateTime`
+ `func GetDuration(data interface{}) neo4j.Duration`
+ `func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue`
+ `func GenerateOrderByExp(source string, dict PropertyDict) string`
+ `func NewPool(driver neo4j.Driver, dial DialHandler) *Pool`

### Methods

+ `func (p *PropertyValue) Destinations() []string`
+ `func (p *PropertyValue) Reverse() bool`
+ `func (p *Pool) Dial(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error)`
+ `func (p *Pool) DialReadMode(bookmarks ...string) (neo4j.Session, error)`
+ `func (p *Pool) DialWriteMode(bookmarks ...string) (neo4j.Session, error)`
+ `func (p *Pool) Target() url.URL`
+ `func (p *Pool) VerifyConnectivity() error`
+ `func (p *Pool) Close() error`
