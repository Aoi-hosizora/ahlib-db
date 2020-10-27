package xneo4j

import (
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xproperty"
	"github.com/Aoi-hosizora/ahlib/xtime"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"
	"time"
)

// Get records from run result, see neo4j.Collect.
//
// 1. result: array of neo4j.Record;
//
// 2. record: array of columns value -> Get(key) / GetByIndex(index);
//
// Example:
//	cypher := "MATCH p = ()-[r :FRIEND]->(n) RETURN r, n"
//	rec, _ := xneo4j.GetRecords(session.Run(cypher, nil)) // slice of neo4j.Record
//	for _, r := range rec { // slice of value (interface{})
//		rel := xneo4j.GetRel(r.Values()[0]) // neo4j.Node
//		node := xneo4j.GetNode(r.Values()[1]) // neo4j.Relationship
//		log.Println(rel.Id(), rel.Type(), node.Id(), node.Props())
//	}
func GetRecords(result neo4j.Result, err error) ([]neo4j.Record, error) {
	records, _, err := GetRecordsWithSummary(result, err)
	return records, err
}

// Get neo4j.Record slice and neo4j.ResultSummary.
func GetRecordsWithSummary(result neo4j.Result, err error) ([]neo4j.Record, neo4j.ResultSummary, error) {
	if err != nil {
		return nil, nil, err
	} else if result == nil {
		return nil, nil, fmt.Errorf("nil neo4j result")
	}

	summary, err := result.Summary()
	if err != nil {
		return nil, nil, err
	}

	rec := make([]neo4j.Record, 0)
	for result.Next() {
		rec = append(rec, result.Record())
	}
	if err := result.Err(); err != nil {
		return nil, nil, err
	}

	return rec, summary, nil
}

func NowLocalDateTime() neo4j.LocalDateTime {
	return neo4j.LocalDateTimeOf(xtime.ToDateTime(time.Now()))
}

func NowLocalTime() neo4j.LocalTime {
	return neo4j.LocalTimeOf(xtime.ToDateTime(time.Now()))
}

func NowDate() neo4j.Date {
	return neo4j.DateOf(xtime.ToDate(time.Now()))
}

func NowTime() neo4j.OffsetTime {
	return neo4j.OffsetTimeOf(xtime.ToDateTime(time.Now()))
}

func GetData(records []neo4j.Record, column int, row int) interface{} {
	return records[column].GetByIndex(row)
}

func GetInteger(data interface{}) int64 {
	return data.(int64)
}

func GetFloat(data interface{}) float64 {
	return data.(float64)
}

func GetString(data interface{}) string {
	return data.(string)
}

func GetBoolean(data interface{}) bool {
	return data.(bool)
}

func GetByteArray(data interface{}) []byte {
	return data.([]byte)
}

func GetList(data interface{}) []interface{} {
	return data.([]interface{})
}

func GetMap(data interface{}) map[string]interface{} {
	return data.(map[string]interface{})
}

func GetNode(data interface{}) neo4j.Node {
	return data.(neo4j.Node)
}

func GetRel(data interface{}) neo4j.Relationship {
	return data.(neo4j.Relationship)
}

func GetPath(data interface{}) neo4j.Path {
	return data.(neo4j.Path)
}

func GetPoint(data interface{}) neo4j.Point {
	return data.(neo4j.Point)
}

func GetDate(data interface{}) neo4j.Date {
	return data.(neo4j.Date)
}

func GetTime(data interface{}) neo4j.OffsetTime {
	return data.(neo4j.OffsetTime)
}

func GetLocalTime(data interface{}) neo4j.LocalTime {
	return data.(neo4j.LocalTime)
}

func GetDateTime(data interface{}) time.Time {
	return data.(time.Time)
}

func GetLocalDateTime(data interface{}) neo4j.LocalDateTime {
	return data.(neo4j.LocalDateTime)
}

func GetDuration(data interface{}) neo4j.Duration {
	return data.(neo4j.Duration)
}

func OrderByFunc(p xproperty.PropertyDict) func(source, parent string) string {
	return func(source, parent string) string {
		result := make([]string, 0)
		if source == "" {
			return ""
		}

		sources := strings.Split(source, ",")
		for _, src := range sources {
			if src == "" {
				continue
			}

			src = strings.TrimSpace(src)
			reverse := strings.HasSuffix(src, " desc") || strings.HasSuffix(src, " DESC")
			src = strings.Split(src, " ")[0]

			dest, ok := p[src]
			if !ok || dest == nil || len(dest.Destinations) == 0 {
				continue
			}

			if dest.Revert {
				reverse = !reverse
			}
			for _, prop := range dest.Destinations {
				prop = parent + "." + prop
				if !reverse {
					prop += " ASC"
				} else {
					prop += " DESC"
				}
				result = append(result, prop)
			}
		}

		return strings.Join(result, ", ")
	}
}

func OrderByFunc2(p xproperty.PropertyDict, v xproperty.VariableDict) func(source string, parents ...string) string {
	return func(source string, parents ...string) string {
		result := make([]string, 0)
		if source == "" {
			return ""
		}

		sources := strings.Split(source, ",")
		for _, src := range sources {
			if src == "" {
				continue
			}

			src = strings.TrimSpace(src)
			reverse := strings.HasSuffix(src, " desc") || strings.HasSuffix(src, " DESC")
			src = strings.Split(src, " ")[0]

			dest, ok := p[src]
			if !ok || dest == nil || len(dest.Destinations) == 0 {
				continue
			}
			if len(v) == 0 {
				return ""
			}

			if dest.Revert {
				reverse = !reverse
			}
			for _, prop := range dest.Destinations {
				idx, ok := v[prop]
				if ok {
					idx-- // 1 -> 0
				} else {
					idx = 0
				}
				if len(parents) < idx {
					idx = 0
				}

				prop = parents[idx] + "." + prop
				if !reverse {
					prop += " ASC"
				} else {
					prop += " DESC"
				}
				result = append(result, prop)
			}
		}

		return strings.Join(result, ", ")
	}
}
