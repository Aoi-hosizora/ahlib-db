package xdbutils_mysql

// DSNFormatter is an interface for types which implement FormatDSN method.
type DSNFormatter interface {
	FormatDSN() string
}

// ParamExtender is an interface for types which implement ExtendParam method.
type ParamExtender interface {
	ExtendParam(others map[string]string) map[string]string
}

// MySQLExtraConfig is an extra configuration for MySQLConfig, can be used to generate extends given param by ExtendParam method. Please visit
// https://github.com/go-sql-driver/mysql#dsn-data-source-name and https://dev.mysql.com/doc/refman/8.0/en/connecting-using-uri-or-key-value-pairs.html
// for more information.
type MySQLExtraConfig struct {
	Charset              string
	Autocommit           string
	TimeZone             string
	TransactionIsolation string
	SQLMode              string
	SysVar               string
}

// ExtendParam extends given param from MySQLExtraConfig.
func (m *MySQLExtraConfig) ExtendParam(others map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range others {
		result[k] = v
	}

	if m.Charset != "" {
		result["charset"] = m.Charset // prefer "utf8mb4,utf8"
	}
	if m.Autocommit != "" {
		result["autocommit"] = m.Autocommit
	}
	if m.TimeZone != "" {
		result["time_zone"] = m.TimeZone
	}
	if m.TransactionIsolation != "" {
		result["transaction_isolation"] = m.TransactionIsolation
	}
	if m.SQLMode != "" {
		result["sql_mode"] = m.SQLMode
	}
	if m.SysVar != "" {
		result["sys_var"] = m.SysVar
	}

	return result
}
