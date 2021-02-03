package xneo4j

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"
	"time"
)

// Example of neo4j.Collect:
// 	cypher := "MATCH p = ()-[r :FRIEND]->(n) RETURN r, n"
// 	records, _ := neo4j.Collect(session.Run(cypher, nil)) // records is a slice of neo4j.Record
// 	for _, record := range records {
// 		// record is a slice of value (interface{}), can be get by Get(key string) or GetByIndex(index int)
// 		rel := xneo4j.GetRel(record.Values()[0]) // neo4j.Relationship
// 		relId, relTyp, relProps := rel.Id(), rel.Type(), rel.Props()
// 		node := xneo4j.GetNode(record.Values()[1]) // neo4j.Node
// 		nodeId, nodeLabels, nodeProps := node.Id(), node.Labels(), node.Props()
// 	}
var _ = neo4j.Collect

// GetByColumnIndex gets data from neo4j.Record-s by given row and column (return list) index, return false when index out of range.
func GetByColumnIndex(records []neo4j.Record, row int, index int) (interface{}, bool) {
	if row >= len(records) {
		return nil, false
	}
	rec := records[row]
	if index >= len(rec.Keys()) {
		return nil, false
	}
	return rec.GetByIndex(index), true
}

// GetByColumnKey gets data from neo4j.Record-s by given row and column (return list) key, return false when index out of range or key is not found.
func GetByColumnKey(records []neo4j.Record, row int, key string) (interface{}, bool) {
	if row >= len(records) {
		return nil, false
	}
	return records[row].Get(key)
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

// GetLocalTime returns neo4j LocalTime value (neo4j.LocalTime) from given data.
func GetLocalTime(data interface{}) neo4j.LocalTime {
	return data.(neo4j.LocalTime)
}

// GetDateTime returns neo4j DateTime value (time.Time) from given data.
func GetDateTime(data interface{}) time.Time {
	return data.(time.Time)
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
type PropertyValue struct {
	destinations []string // mapping destinations
	reverse      bool     // reverse order mapping
}

// NewPropertyValue creates a PropertyValue by given reverse and destinations.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	finalDestinations := make([]string, 0, len(destinations))
	for _, dest := range destinations {
		dest = strings.TrimSpace(dest)
		if dest != "" {
			finalDestinations = append(finalDestinations, dest) // filter empty destination
		}
	}
	return &PropertyValue{reverse: reverse, destinations: finalDestinations}
}

// Destinations returns the destinations of PropertyValue.
func (p *PropertyValue) Destinations() []string {
	return p.destinations
}

// Reverse returns the reverse of PropertyValue.
func (p *PropertyValue) Reverse() bool {
	return p.reverse
}

// PropertyDict represents a DTO-PO PropertyValue dictionary, used in GenerateOrderByExp.
type PropertyDict map[string]*PropertyValue

// GenerateOrderByExp returns a generated orderBy expresion by given source dto order string (split by ",") and PropertyDict.
func GenerateOrderByExp(source string, dict PropertyDict) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return ""
	}

	result := make([]string, 0)
	for _, src := range strings.Split(source, ",") {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}
		srcSp := strings.Split(src, " ") // xxx / yyy asc / zzz desc
		if len(srcSp) > 2 {
			continue
		}

		src = srcSp[0]
		desc := len(srcSp) == 2 && strings.ToUpper(srcSp[1]) == "DESC"
		value, ok := dict[src] // property mapping rule
		if !ok || value == nil || len(value.destinations) == 0 {
			continue
		}

		if value.reverse {
			desc = !desc
		}
		for _, prop := range value.destinations {
			direction := " ASC"
			if !desc {
				direction = " DESC"
			}
			result = append(result, prop+direction)
		}
	}

	return strings.Join(result, ", ")
}
