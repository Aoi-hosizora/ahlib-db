package xneo4j

import (
	"github.com/Aoi-hosizora/ahlib-db/internal/orderby"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"time"
)

// P represents the cypher parameter type, equals to "map[string]interface{}".
// Example:
// 	session.Run(`MATCH (n {id: $id}) RETURN n`, xneo4j.P{"id": 2})
type P map[string]interface{}

// Collect loops through the result stream, collects records into a slice and returns the resulting slice with result summary.
//
// Cypher manual refers to https://neo4j.com/docs/cypher-manual/3.5/syntax/.
// Neo4j go driver refers to https://github.com/neo4j/neo4j-go-driver/tree/1.8.
//
// Example:
// 	cypher := "MATCH p = ()-[r :FRIEND]->(n) RETURN r, n"
// 	records, summary, err := xneo4j.Collect(session.Run(cypher, nil)) // err contains connect and execute error
// 	for _, record := range records { // records is a slice of neo4j.Record
// 		// record is the returned values, each value can be get by `Get` or `GetByIndex` methods
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
func GetPoint(data interface{}) *neo4j.Point {
	if p, ok := data.(neo4j.Point); ok {
		return &p
	}
	return data.(*neo4j.Point)
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

// PropertyValue represents a PO entity's property mapping rule.
type PropertyValue = orderby.PropertyValue

// PropertyDict represents a DTO-PO PropertyValue dictionary, used in GenerateOrderByExp.
type PropertyDict = orderby.PropertyDict

// NewPropertyValue creates a PropertyValue by given reverse and destinations.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	return orderby.NewPropertyValue(reverse, destinations...)
}

// GenerateOrderByExp returns a generated orderBy expresion by given source dto order string (split by ",", such as "name desc, age asc") and PropertyDict.
// The generated expression is in mysql-sql and neo4j-cypher style, that is "xx ASC", "xx DESC".
func GenerateOrderByExp(source string, dict PropertyDict) string {
	return orderby.GenerateOrderByExp(source, dict)
}
