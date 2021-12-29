package internal

import (
	"strings"
)

// PropertyValue represents a PO entity's property mapping rule.
type PropertyValue struct {
	// destinations represents mapping property destination array.
	//
	// If using sql, use `property_name` directly.
	// If using cypher, use `returned_name.property_name`.
	destinations []string

	// reverse represents the switcher for revert of order
	reverse bool
}

// PropertyDict represents a DTO-PO PropertyValue dictionary, used in GenerateOrderByExp.
type PropertyDict map[string]*PropertyValue

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

// GenerateOrderByExp returns a generated orderBy expression by given source dto order string (split by ",", such as "name desc,age asc") and PropertyDict.
// The generated expression is in mysql-sql and neo4j-cypher style, that is "xx ASC", "xx DESC".
func GenerateOrderByExp(source string, dict PropertyDict) string {
	source = strings.TrimSpace(source)
	if source == "" || len(dict) == 0 {
		return ""
	}

	result := make([]string, 0)
	for _, src := range strings.Split(source, ",") {
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}

		srcSp := strings.Split(src, " ") // xxx / yyy asc / zzz desc
		src = srcSp[0]
		desc := len(srcSp) >= 2 && strings.ToUpper(srcSp[1]) == "DESC"
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
