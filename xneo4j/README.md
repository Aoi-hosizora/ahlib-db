# xneo4j

### Functions

#### Logger

+ `type LogrusNeo4j struct {}`
+ `NewLogrusNeo4j(session neo4j.Session, logger *logrus.Logger, logMode bool) *LogrusNeo4j`
+ `type LoggerNeo4j struct {}`
+ `NewLoggerNeo4j(session neo4j.Session, logger *log.Logger, logMode bool) *LoggerNeo4j`

#### Helper

+ `type DialFunc func(driver neo4j.Driver) (neo4j.Session, error)`
+ `NewNeo4jPool(driver neo4j.Driver, dial DialFunc) *Neo4jPool`
+ `(n *Pool) Get(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error)`
+ `(n *Pool) GetWriteMode(bookmarks ...string) (neo4j.Session, error)`
+ `(n *Pool) GetReadMode(bookmarks ...string) (neo4j.Session, error)`
+ `GetRecords(result neo4j.Result) ([]neo4j.Record, error)`
+ `GetRecordsWithSummary(result neo4j.Result, err error) ([]neo4j.Record, neo4j.ResultSummary, error)`
+ `NowLocalDateTime() neo4j.LocalDateTime`
+ `NowLocalTime() neo4j.LocalTime`
+ `NowDate() neo4j.Date`
+ `NowTime() neo4j.OffsetTime`
+ `GetData(records []neo4j.Record, column int, row int) interface{}`
+ `GetInteger(data interface{}) int64`
+ `GetFloat(data interface{}) float64`
+ `GetString(data interface{}) string`
+ `GetBoolean(data interface{}) bool`
+ `GetByteArray(data interface{}) []byte`
+ `GetList(data interface{}) []interface{}`
+ `GetMap(data interface{}) map[string]interface{}`
+ `GetNode(data interface{}) neo4j.Node`
+ `GetRel(data interface{}) neo4j.Relationship`
+ `GetPath(data interface{}) neo4j.Path`
+ `GetPoint(data interface{}) neo4j.Point`
+ `GetDate(data interface{}) neo4j.Date`
+ `GetTime(data interface{}) neo4j.OffsetTime`
+ `GetLocalTime(data interface{}) neo4j.LocalTime`
+ `GetDateTime(data interface{}) time.Time`
+ `GetLocalDateTime(data interface{}) neo4j.LocalDateTime`
+ `GetDuration(data interface{}) neo4j.Duration`
+ `OrderByFunc(p xproperty.PropertyDict) func(source, parent string) string`
+ `OrderByFunc2(p xproperty.PropertyDict, v xproperty.VariableDict) func(source string, parents ...string) string`
