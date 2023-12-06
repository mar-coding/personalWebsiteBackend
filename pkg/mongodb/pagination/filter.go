package pagination

// Text filter text fields from pagination
func Text(field string, operator FilterOperator) Columned {
	return newColumnInfo(field, operator, TEXT)
}

// Int filter int fields from pagination with custom operator
func Int(field string, operator FilterOperator) Columned {
	return newColumnInfo(field, operator, INT)
}

// Bool filter bool fields from pagination
func Bool(field string) Columned {
	return newColumnInfo(field, "=", BOOL)
}

// Float filter float64 fields from pagination
func Float(field string, operator FilterOperator) Columned {
	return newColumnInfo(field, operator, FLOAT64)
}

// ObjectID filter objectID fields from pagination with operator
func ObjectID(field string, operator FilterOperator) Columned {
	return newColumnInfoWithExportSettings(field, operator, OBJECT_ID, &ExportSettings{Show: false})
}

// Date filter datetime from pagination
func Date(field string, operator FilterOperator) Columned {
	return newColumnInfo(field, operator, DATE)
}
