# xdbutils_orderby

## Dependencies

+ None

## Documents

### Types

+ `type PropertyValue struct`
+ `type PropertyDict map`
+ `type OrderByOption func`

### Variables

+ None

### Constants

+ None

### Functions

+ `func NewPropertyValue(reverse bool, destinations ...string) *PropertyValue`
+ `func WithSourceSeparator(separator string) OrderByOption`
+ `func WithTargetSeparator(separator string) OrderByOption`
+ `func WithSourceProcessor(processor func(source string) (field string, asc bool)) OrderByOption`
+ `func WithTargetProcessor(processor func(destination string, asc bool) (target string)) OrderByOption`
+ `func GenerateOrderByExpr(querySource string, dict PropertyDict, options ...OrderByOption) string`

### Methods

+ None
