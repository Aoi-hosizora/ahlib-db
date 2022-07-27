# xdbutils_sqlite

## Dependencies

+ None

## Documents

### Types

+ `type ErrNo int`
+ `type ErrNoExtended int`
+ `type SQLiteConfig struct`

### Variables

+ None

### Constants

+ ...

### Functions

+ `func CheckSQLiteErrorExtendedCodeByReflect(err error, code int) bool`

### Methods

+ `func (err ErrNo) Extend(by int) ErrNoExtended`
+ `func (s *SQLiteConfig) FormatDSN() string`
