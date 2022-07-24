package xneo4j

import (
	"github.com/Aoi-hosizora/ahlib-db/xdbutils/xdbutils_orderby"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"time"
)

// Cypher manual 3.5 can refer to https://neo4j.com/docs/cypher-manual/3.5/syntax/.
// Cypher manual 4.0 can refer to https://neo4j.com/docs/cypher-manual/4.0/syntax/.
// Neo4j go driver 1.x can refer to https://github.com/neo4j/neo4j-go-driver/tree/1.8.
// Neo4j go driver 4.x can refer to https://github.com/neo4j/neo4j-go-driver/master.

// P is a cypher parameters type, is always same as `map[string]interface{}`.
//
// Example:
// 	session.Run(`MATCH (n {id: $id}) RETURN n`, xneo4j.P{"id": 2})
type P = map[string]interface{}

// =========================
// collect and get functions
// =========================

// Collect loops through the result stream, collects records and returns the neo4j.Record slice with neo4j.ResultSummary.
//
// Example:
// 	cypher := "MATCH p = ()-[r :FRIEND]->(n) RETURN r, n"
// 	records, summary, err := xneo4j.Collect(session.Run(cypher, nil)) // err contains the connection and execution error
// 	for _, record := range records { // records is a slice of neo4j.Record
// 		// record is returned values, each value can be got by `Get` or `GetByIndex` methods
// 		rel := xneo4j.GetRel(record.GetByIndex(0))   // neo4j.Relationship
// 		node := xneo4j.GetNode(record.GetByIndex(1)) // neo4j.Node
// 	}
func Collect(result neo4j.Result, err error) ([]neo4j.Record, neo4j.ResultSummary, error) {
	if err != nil {
		return nil, nil, err // failed to connect
	}
	summary, err := result.Summary()
	if err != nil {
		return nil, nil, err // failed to execute
	}
	records, err := neo4j.Collect(result, err)
	if err != nil {
		return nil, nil, err // ...
	}
	return records, summary, nil
}

// GetInteger returns neo4j Integer value (int64) from given data.
func GetInteger(data interface{}) int64 {
	return data.(int64)
}

// GetFloat returns neo4j Float value (float64) from given data.
func GetFloat(data interface{}) float64 {
	return data.(float64)
}

// GetString returns neo4j String value (string) from given data.
func GetString(data interface{}) string {
	return data.(string)
}

// GetBoolean returns neo4j Boolean value (bool) from given data.
func GetBoolean(data interface{}) bool {
	return data.(bool)
}

// GetByteArray returns neo4j ByteArray value ([]byte) from given data.
func GetByteArray(data interface{}) []byte {
	return data.([]byte)
}

// GetList returns neo4j List value ([]interface{}) from given data.
func GetList(data interface{}) []interface{} {
	return data.([]interface{})
}

// GetMap returns neo4j Map value (map[string]interface{}) from given data.
func GetMap(data interface{}) map[string]interface{} {
	return data.(map[string]interface{})
}

// GetNode returns neo4j Node value (neo4j.Node) from given data.
func GetNode(data interface{}) neo4j.Node {
	return data.(neo4j.Node)
}

// GetRel returns neo4j Relationship value (neo4j.Relationship) from given data.
func GetRel(data interface{}) neo4j.Relationship {
	return data.(neo4j.Relationship)
}

// GetPath returns neo4j Path value (neo4j.Path) from given data.
func GetPath(data interface{}) neo4j.Path {
	return data.(neo4j.Path)
}

// GetPoint returns neo4j Point value (neo4j.Point) from given data.
func GetPoint(data interface{}) neo4j.Point {
	return data.(neo4j.Point)
}

// GetDate returns neo4j Date value (neo4j.Date) from given data.
func GetDate(data interface{}) neo4j.Date {
	return data.(neo4j.Date)
}

// GetTime returns neo4j Time value (neo4j.OffsetTime) from given data.
func GetTime(data interface{}) neo4j.OffsetTime {
	return data.(neo4j.OffsetTime)
}

// GetDateTime returns neo4j DateTime value (time.Time) from given data.
func GetDateTime(data interface{}) time.Time {
	return data.(time.Time)
}

// GetLocalTime returns neo4j LocalTime value (neo4j.LocalTime) from given data.
func GetLocalTime(data interface{}) neo4j.LocalTime {
	return data.(neo4j.LocalTime)
}

// GetLocalDateTime returns neo4j LocalDateTime value (neo4j.LocalDateTime) from given data.
func GetLocalDateTime(data interface{}) neo4j.LocalDateTime {
	return data.(neo4j.LocalDateTime)
}

// GetDuration returns neo4j Duration value (neo4j.Duration) from given data.
func GetDuration(data interface{}) neo4j.Duration {
	return data.(neo4j.Duration)
}

// ========
// order by
// ========

// PropertyValue represents database single entity's property mapping rule, is used in GenerateOrderByExpr.
type PropertyValue = xdbutils_orderby.PropertyValue

// PropertyDict is used to store PropertyValue-s for data transfer object (dto) to entity's property mapping rule, is used in GenerateOrderByExpr.
type PropertyDict = xdbutils_orderby.PropertyDict

// OrderByOption represents an option type for GenerateOrderByExpr's option, can be created by WithXXX functions.
type OrderByOption = xdbutils_orderby.OrderByOption

// WithSourceSeparator creates an OrderByOption to specify the source order-by expression fields separator, defaults to ",".
func WithSourceSeparator(separator string) OrderByOption {
	return xdbutils_orderby.WithSourceSeparator(separator)
}

// WithTargetSeparator creates an OrderByOption to specify the target order-by expression fields separator, defaults to ", ".
func WithTargetSeparator(separator string) OrderByOption {
	return xdbutils_orderby.WithTargetSeparator(separator)
}

// WithSourceProcessor creates an OrderByOption to specify the source processor for extracting field name and ascending flag from given source,
// defaults to use the "field asc" or "field desc" format (case-insensitive) to extract information.
func WithSourceProcessor(processor func(source string) (field string, asc bool)) OrderByOption {
	return xdbutils_orderby.WithSourceProcessor(processor)
}

// WithTargetProcessor creates an OrderByOption to specify the target processor for combining field name and ascending flag to target expression,
// defaults to generate the target with "destination ASC" or "destination DESC" format.
func WithTargetProcessor(processor func(destination string, asc bool) (target string)) OrderByOption {
	return xdbutils_orderby.WithTargetProcessor(processor)
}

// NewPropertyValue creates a PropertyValue by given reverse and destinations, is used to describe database single entity's property mapping rule.
//
// Here:
// 1. `destinations` represents mapping property destination list, use `property_name` directly for sql, use `returned_name.property_name` for cypher.
// 2. `reverse` represents the flag whether you need to revert the order or not.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	return xdbutils_orderby.NewPropertyValue(reverse, destinations...)
}

// GenerateOrderByExpr returns a generated order-by expression by given order-by query source string (such as "name desc, age asc") and PropertyDict,
// with some OrderByOption-s. The generated expression will be in mysql-sql (such as "xxx ASC") or neo4j-cypher style (such as "xxx.yyy DESC").
//
// Example:
// 	dict := PropertyDict{
// 		"uid":  NewPropertyValue(false, "p.uid"),
// 		"name": NewPropertyValue(false, "p.firstname", "p.lastname"),
// 		"age":  NewPropertyValue(true, "u.birthday"),
// 	}
// 	_ = GenerateOrderByExpr(`uid, age desc`, dict) // => p.uid ASC, u.birthday ASC
// 	_ = GenerateOrderByExpr(`age, username desc`, dict) // => u.birthday DESC, p.firstname DESC, p.lastname DESC
func GenerateOrderByExpr(querySource string, dict PropertyDict, options ...OrderByOption) string {
	return xdbutils_orderby.GenerateOrderByExpr(querySource, dict, options...)
}

// ====================
// neo4j config options
// ====================

// DriverOption represents an option type for neo4j.NewDriver's option, can be created by WithXXX functions.
type DriverOption func(*neo4j.Config)

// WithEncrypted returns a neo4j.Config option function to specify encrypted flag for neo4j.Driver, defaults to true.
func WithEncrypted(encrypted bool) DriverOption {
	return func(config *neo4j.Config) {
		config.Encrypted = encrypted
	}
}

// WithTrustStrategy returns a neo4j.Config option function to specify trust strategy for neo4j.Driver, defaults to neo4j.TrustAny(false).
func WithTrustStrategy(e neo4j.TrustStrategy) DriverOption {
	return func(config *neo4j.Config) {
		config.TrustStrategy = e
	}
}

// WithLog returns a neo4j.Config option function to specify log function for neo4j.Driver, defaults to neo4j.NoOpLogger.
func WithLog(l neo4j.Logging) DriverOption {
	return func(config *neo4j.Config) {
		config.Log = l
	}
}

// WithAddressResolver returns a neo4j.Config option function to specify address resolver for neo4j.Driver, defaults to nil.
func WithAddressResolver(resolver neo4j.ServerAddressResolver) DriverOption {
	return func(config *neo4j.Config) {
		config.AddressResolver = resolver
	}
}

// WithMaxTransactionRetryTime returns a neo4j.Config option function to specify max transaction retry time for neo4j.Driver, defaults to 30s.
func WithMaxTransactionRetryTime(t time.Duration) DriverOption {
	return func(config *neo4j.Config) {
		config.MaxTransactionRetryTime = t
	}
}

// WithMaxConnectionPoolSize returns a neo4j.Config option function to specify max connection pool size for neo4j.Driver, defaults to 100.
func WithMaxConnectionPoolSize(size int) DriverOption {
	return func(config *neo4j.Config) {
		config.MaxConnectionPoolSize = size
	}
}

// WithMaxConnectionLifetime returns a neo4j.Config option function to specify max connection lifetime for neo4j.Driver, defaults to 1h.
func WithMaxConnectionLifetime(t time.Duration) DriverOption {
	return func(config *neo4j.Config) {
		config.MaxConnectionLifetime = t
	}
}

// WithConnectionAcquisitionTimeout returns a neo4j.Config option function to specify connection acquisition timeout for neo4j.Driver, defaults to 1min.
func WithConnectionAcquisitionTimeout(t time.Duration) DriverOption {
	return func(config *neo4j.Config) {
		config.ConnectionAcquisitionTimeout = t
	}
}

// WithSocketConnectTimeout returns a neo4j.Config option function to specify socket connect timeout for neo4j.Driver, defaults to 5s.
func WithSocketConnectTimeout(t time.Duration) DriverOption {
	return func(config *neo4j.Config) {
		config.SocketConnectTimeout = t
	}
}

// WithSocketKeepalive returns a neo4j.Config option function to specify socket keepalive flag for neo4j.Driver, defaults to true.
func WithSocketKeepalive(keepalive bool) DriverOption {
	return func(config *neo4j.Config) {
		config.SocketKeepalive = keepalive
	}
}
