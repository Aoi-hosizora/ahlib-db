package xdbutils

import (
	"strings"
)

// PropertyValue is a struct type of database entity's property mapping rule, used in GenerateOrderByExpr.
type PropertyValue struct {
	destinations []string
	reverse      bool
}

// Destinations returns the destinations of PropertyValue.
func (p *PropertyValue) Destinations() []string {
	return p.destinations
}

// Reverse returns the reverse of PropertyValue.
func (p *PropertyValue) Reverse() bool {
	return p.reverse
}

// PropertyDict is a dictionary type to store pairs from data transfer object to database entity's PropertyValue, used in GenerateOrderByExpr.
type PropertyDict map[string]*PropertyValue

// NewPropertyValue creates a PropertyValue by given reverse and destinations, used to describe database entity's property mapping rule.
//
// Here:
// 1. `destinations` represents mapping property destination array, use `property_name` directly for sql, use `returned_name.property_name` for cypher.
// 2. `reverse` represents the flag whether you need to revert the order or not.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	target := make([]string, 0, len(destinations))
	for _, d := range destinations {
		d = strings.TrimSpace(d)
		if len(d) > 0 {
			target = append(target, d)
		}
	}
	return &PropertyValue{reverse: reverse, destinations: target}
}

// GenerateOrderByExpr returns a generated order-by expression by given source (query string) order string (such as "name desc, age asc") and PropertyDict.
// The generated expression is in mysql-sql or neo4j-cypher style (such as "xxx ASC" or "xxx.yyy DESC").
//
// SQL Example:
// 	dict := PropertyDict{
// 		"uid":  NewPropertyValue(false, "uid"),
// 		"name": NewPropertyValue(false, "firstname", "lastname"),
// 		"age":  NewPropertyValue(true, "birthday"),
// 	}
// 	_ = GenerateOrderByExpr(`uid, age desc`, dict) // => uid ASC, birthday ASC
// 	_ = GenerateOrderByExpr(`age, username desc`, dict) // => birthday DESC, firstname DESC, lastname DESC
//
// Cypher Example:
// 	dict := PropertyDict{
// 		"uid":  NewPropertyValue(false, "p.uid"),
// 		"name": NewPropertyValue(false, "p.firstname", "p.lastname"),
// 		"age":  NewPropertyValue(true, "u.birthday"),
// 	}
// 	_ = GenerateOrderByExpr(`uid, age desc`, dict) // => p.uid ASC, u.birthday ASC
// 	_ = GenerateOrderByExpr(`age, username desc`, dict) // => u.birthday DESC, p.firstname DESC, p.lastname DESC
func GenerateOrderByExpr(source string, dict PropertyDict) string {
	source = strings.TrimSpace(source)
	if len(source) == 0 || len(dict) == 0 {
		return ""
	}

	sources := strings.Split(source, ",")
	result := make([]string, 0, len(sources))
	for _, src := range sources {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}

		srcSp := strings.Split(src, " ") // xxx / yyy asc / zzz desc
		src = srcSp[0]
		desc := len(srcSp) >= 2 && strings.ToLower(srcSp[1]) == "desc"
		value, ok := dict[src] // property mapping rule
		if !ok || value == nil || len(value.destinations) == 0 {
			continue
		}

		if value.reverse {
			desc = !desc
		}
		for _, prop := range value.destinations {
			if !desc {
				prop += " ASC"
			} else {
				prop += " DESC"
			}
			result = append(result, prop)
		}
	}

	return strings.Join(result, ", ")
}
