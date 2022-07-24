# xneo4j

## Dependencies

+ github.com/Aoi-hosizora/ahlib
+ github.com/neo4j/neo4j-go-driver
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type P map`
+ `type PropertyValue struct`
+ `type PropertyDict map`
+ `type OrderByOption func`
+ `type DriverOption func`
+ `type DialHandler func`
+ `type Pool struct`
+ `type LoggerOption func`
+ `type LogrusLogger struct`
+ `type StdLogger struct`
+ `type LoggerParam struct`

### Variables

+ `var FormatLoggerFunc func`
+ `var FieldifyLoggerFunc func`

### Constants

+ `const DefaultDatabase string`

### Functions

+ `func Collect(result neo4j.Result, err error) ([]neo4j.Record, neo4j.ResultSummary, error)`
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
+ `func GetDateTime(data interface{}) time.Time`
+ `func GetLocalTime(data interface{}) neo4j.LocalTime`
+ `func GetLocalDateTime(data interface{}) neo4j.LocalDateTime`
+ `func GetDuration(data interface{}) neo4j.Duration`
+ `func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue`
+ `func WithSourceSeparator(separator string) OrderByOption`
+ `func WithTargetSeparator(separator string) OrderByOption`
+ `func WithSourceProcessor(processor func(source string) (field string, asc bool)) OrderByOption`
+ `func WithTargetProcessor(processor func(destination string, asc bool) (target string)) OrderByOption`
+ `func GenerateOrderByExpr(querySource string, dict PropertyDict, options ...OrderByOption) string`
+ `func WithEncrypted(encrypted bool) DriverOption`
+ `func WithTrustStrategy(e neo4j.TrustStrategy) DriverOption`
+ `func WithLog(l neo4j.Logging) DriverOption`
+ `func WithAddressResolver(resolver neo4j.ServerAddressResolver) DriverOption`
+ `func WithMaxTransactionRetryTime(t time.Duration) DriverOption`
+ `func WithMaxConnectionPoolSize(size int) DriverOption`
+ `func WithMaxConnectionLifetime(t time.Duration) DriverOption`
+ `func WithConnectionAcquisitionTimeout(t time.Duration) DriverOption`
+ `func WithSocketConnectTimeout(t time.Duration) DriverOption`
+ `func WithSocketKeepalive(keepalive bool) DriverOption`
+ `func NewPool(driver neo4j.Driver, dial DialHandler) *Pool`
+ `func WithLogErr(log bool) LoggerOption`
+ `func WithLogCypher(log bool) LoggerOption`
+ `func WithCounterFields(flag bool) LoggerOption`
+ `func WithSkip(skip int) LoggerOption`
+ `func WithSlowThreshold(threshold time.Duration) LoggerOption`
+ `func EnableLogger()`
+ `func DisableLogger()`
+ `func NewLogrusLogger(session neo4j.Session, logger *logrus.Logger, options ...LoggerOption) *LogrusLogger`
+ `func NewStdLogger(session neo4j.Session, logger logrus.StdLogger, options ...LoggerOption) *StdLogger`

### Methods

+ `func (p *PropertyValue) Destinations() []string`
+ `func (p *PropertyValue) Reverse() bool`
+ `func (p *Pool) Dial(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error)`
+ `func (p *Pool) NewSession(config neo4j.SessionConfig) (neo4j.Session, error)`
+ `func (p *Pool) Session(mode neo4j.AccessMode, bookmarks ...string) (neo4j.Session, error)`
+ `func (p *Pool) DialReadMode(bookmarks ...string) (neo4j.Session, error)`
+ `func (p *Pool) DialWriteMode(bookmarks ...string) (neo4j.Session, error)`
+ `func (p *Pool) Driver() neo4j.Driver`
+ `func (p *Pool) Target() url.URL`
+ `func (p *Pool) VerifyConnectivity() error`
+ `func (p *Pool) Close() error`
+ `func (l *LogrusLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error)`
+ `func (l *StdLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error)`
