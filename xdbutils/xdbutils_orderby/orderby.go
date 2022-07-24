package xdbutils_orderby

import (
	"strings"
)

// PropertyValue represents database single entity's property mapping rule, is used in GenerateOrderByExpr.
type PropertyValue struct {
	destinations []string
	reverse      bool
}

// Destinations returns the destinations of PropertyValue.
func (p *PropertyValue) Destinations() []string {
	return p.destinations
}

// Reverse returns the reverse flag of PropertyValue.
func (p *PropertyValue) Reverse() bool {
	return p.reverse
}

// PropertyDict is used to store PropertyValue-s for data transfer object (dto) to entity's property mapping rule, is used in GenerateOrderByExpr.
type PropertyDict map[string]*PropertyValue

// NewPropertyValue creates a PropertyValue by given reverse and destinations, is used to describe database single entity's property mapping rule.
//
// Here:
// 1. `destinations` represents mapping property destination list, use `property_name` directly for sql, use `returned_name.property_name` for cypher.
// 2. `reverse` represents the flag whether you need to revert the order or not.
func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue {
	final := make([]string, 0, len(destinations))
	for _, d := range destinations {
		d = strings.TrimSpace(d)
		if len(d) > 0 {
			final = append(final, d)
		}
	}
	return &PropertyValue{reverse: reverse, destinations: final}
}

// orderByOptions is a type of GenerateOrderByExpr's option, each field can be set by OrderByOption function type.
type orderByOptions struct {
	sourceSeparator string
	targetSeparator string
	sourceProcessor func(source string) (field string, asc bool)
	targetProcessor func(destination string, asc bool) (target string)
}

// OrderByOption represents an option type for GenerateOrderByExpr's option, can be created by WithXXX functions.
type OrderByOption func(*orderByOptions)

// WithSourceSeparator creates an OrderByOption to specify the source order-by expression fields separator, defaults to ",".
func WithSourceSeparator(separator string) OrderByOption {
	return func(o *orderByOptions) {
		o.sourceSeparator = separator // no trim
	}
}

// WithTargetSeparator creates an OrderByOption to specify the target order-by expression fields separator, defaults to ", ".
func WithTargetSeparator(separator string) OrderByOption {
	return func(o *orderByOptions) {
		o.targetSeparator = separator // no trim
	}
}

// WithSourceProcessor creates an OrderByOption to specify the source processor for extracting field name and ascending flag from given source,
// defaults to use the "field asc" or "field desc" format (case-insensitive) to extract information.
func WithSourceProcessor(processor func(source string) (field string, asc bool)) OrderByOption {
	return func(o *orderByOptions) {
		o.sourceProcessor = processor
	}
}

// WithTargetProcessor creates an OrderByOption to specify the target processor for combining field name and ascending flag to target expression,
// defaults to generate the target with "destination ASC" or "destination DESC" format.
func WithTargetProcessor(processor func(destination string, asc bool) (target string)) OrderByOption {
	return func(o *orderByOptions) {
		o.targetProcessor = processor
	}
}

// defaultSourceProcessor is the default source processor, for extracting field name and ascending flag from given source.
func defaultSourceProcessor(source string) (field string, asc bool) {
	sp := strings.Split(source, " ") // xxx / yyy asc / zzz desc
	desc := len(sp) >= 2 && strings.ToLower(strings.TrimSpace(sp[1])) == "desc"
	return sp[0], !desc
}

// defaultTargetProcessor is the default target processor, for combining field name and ascending flag to target expression.
func defaultTargetProcessor(destination string, asc bool) (target string) {
	if asc {
		return destination + " ASC" // xxx ASC / yyy.zzz ASC
	}
	return destination + " DESC"
}

// buildOrderByOptions creates a orderByOptions with given OrderByOption-s.
func buildOrderByOptions(options []OrderByOption) *orderByOptions {
	opt := &orderByOptions{}
	for _, o := range options {
		if o != nil {
			o(opt)
		}
	}
	if opt.sourceSeparator == "" {
		opt.sourceSeparator = ","
	}
	if opt.targetSeparator == "" {
		opt.targetSeparator = ", "
	}
	if opt.sourceProcessor == nil {
		opt.sourceProcessor = defaultSourceProcessor
	}
	if opt.targetProcessor == nil {
		opt.targetProcessor = defaultTargetProcessor
	}
	return opt
}

// GenerateOrderByExpr returns a generated order-by expression by given order-by query source string (such as "name desc, age asc") and PropertyDict,
// with some OrderByOption-s. The generated expression will be in mysql-sql (such as "xxx ASC") or neo4j-cypher style (such as "xxx.yyy DESC").
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
func GenerateOrderByExpr(querySource string, dict PropertyDict, options ...OrderByOption) string {
	opt := buildOrderByOptions(options)
	querySource = strings.TrimSpace(querySource)
	if len(querySource) == 0 || len(dict) == 0 {
		return ""
	}

	sources := strings.Split(querySource, opt.sourceSeparator)
	targets := make([]string, 0, len(sources))
	for _, source := range sources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}

		field, asc := opt.sourceProcessor(source)
		rule, ok := dict[field]
		if !ok || rule == nil || len(rule.destinations) == 0 {
			continue
		}

		if rule.reverse {
			asc = !asc
		}
		for _, destination := range rule.destinations {
			target := opt.targetProcessor(destination, asc)
			if target = strings.TrimSpace(target); target != "" {
				targets = append(targets, target)
			}
		}
	}

	return strings.Join(targets, opt.targetSeparator)
}
