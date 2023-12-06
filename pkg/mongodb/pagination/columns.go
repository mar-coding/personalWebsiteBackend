package pagination

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	"time"
)

type (
	FilterOperator string
	ColumnType     string
)

const (
	Greater      FilterOperator = ">"
	GreaterEqual                = ">="
	Less                        = "<"
	LessEqual                   = "<="
	Equal                       = "="
	Between                     = "between"
	In                          = "in"
	NotIn                       = "not in"
	Like                        = "like"
	Is                          = "is"
)

const (
	TEXT      ColumnType = "text"
	DATE                 = "date"
	INT                  = "int"
	BOOL                 = "bool"
	FLOAT64              = "float64"
	OBJECT_ID            = "objectID"

	NONE = "-"
)

type Columned interface {
	none()

	getFilterVal(filterVal string) interface{}
}

type ColumnInfo struct {
	Column           string
	ColumnAlias      string
	FilterOperator   FilterOperator
	DataType         ColumnType
	PrepareValueFunc func(interface{}) interface{}
	ExportSettings   *ExportSettings
}

type ExportSettings struct {
	Show bool
}

func (*ColumnInfo) none() {}

func newColumnInfo(column string, filterOperator FilterOperator, columnType ColumnType) *ColumnInfo {
	return newColumnInfoWithExportSettings(column, filterOperator, columnType, &ExportSettings{Show: true})
}

func newColumnInfoWithExportSettings(column string, filterOperator FilterOperator, columnType ColumnType, exportSettings *ExportSettings) *ColumnInfo {
	return &ColumnInfo{
		Column:         column,
		FilterOperator: filterOperator,
		DataType:       columnType,
		ExportSettings: exportSettings,
	}
}

func (columnInfo *ColumnInfo) setPrepareValueFunc(prepareValueFunc func(interface{}) interface{}) *ColumnInfo {
	columnInfo.PrepareValueFunc = prepareValueFunc
	return columnInfo
}
func (columnInfo *ColumnInfo) hideInExport() *ColumnInfo {
	columnInfo.ExportSettings.Show = false
	return columnInfo
}

func (columnInfo *ColumnInfo) getFilterVal(filterVal string) interface{} {
	switch columnInfo.FilterOperator {
	case Greater:
		return bson.M{"$gt": columnInfo.convertValueToDataType(filterVal)}
	case GreaterEqual:
		return bson.M{"$gte": columnInfo.convertValueToDataType(filterVal)}
	case Less:
		return bson.M{"$lt": columnInfo.convertValueToDataType(filterVal)}
	case LessEqual:
		return bson.M{"$lte": columnInfo.convertValueToDataType(filterVal)}
	case Equal:
		return columnInfo.convertValueToDataType(filterVal)
	case Between:
		values := columnInfo.convertValueToArrayDataType(filterVal)
		if len(values) == 1 || values[1] == nil {
			return bson.M{"$gte": values[0]}
		} else if values[0] == nil {
			return bson.M{"$lte": values[1]}
		} else {
			return bson.M{
				"$gte": values[0],
				"$lte": values[1],
			}
		}
	case In:
		values := columnInfo.convertValueToArrayDataType(filterVal)
		if len(values) == 1 {
			return values[0]
		} else {
			return bson.M{"$in": values}
		}
	case NotIn:
		values := columnInfo.convertValueToArrayDataType(filterVal)
		if len(values) == 1 {
			return bson.M{"$ne": values[0]}
		} else {
			return bson.M{"$nin": values}
		}
	case Like:
		return primitive.Regex{Pattern: filterVal, Options: "i"}
	case Is:
		if filterVal == "0" {
			return bson.TypeNull
		}
		return bson.M{"$ne": bson.TypeNull}
	case NONE:
		break
	}
	return nil
}

func (columnInfo *ColumnInfo) getStringValue(value string) string {
	switch columnInfo.DataType {
	case DATE:
		return parseDateTime(value).Format(time.RFC3339)
	case "":
	case TEXT:
		return value
	}
	return value
}

func (columnInfo *ColumnInfo) convertValueToDataType(value string) interface{} {
	switch columnInfo.DataType {
	case DATE:
		return parseDateTime(value)
	case OBJECT_ID:
		objectId, _ := primitive.ObjectIDFromHex(value)
		return objectId
	case INT:
		if value == "" {
			return nil
		}
		i, _ := strconv.Atoi(value)
		return i
	case FLOAT64:
		if value == "" {
			return nil
		}
		i, _ := strconv.ParseFloat(value, 64)
		return i
	case BOOL:
		if value == "" {
			return nil
		}
		i, _ := strconv.ParseBool(value)
		if !i {
			return bson.M{"$in": bson.A{nil, false}}
		}
		return i
	}
	return value
}

func (columnInfo *ColumnInfo) convertValueToArrayDataType(value string) []interface{} {
	strValues := strings.Split(strings.TrimSpace(value), ",")
	values := []interface{}{}
	for _, strValue := range strValues {
		values = append(values, columnInfo.convertValueToDataType(strValue))
	}
	return values
}

func parseDateTime(datetime string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, datetime)
	if err != nil {
		t, err = time.Parse("2006-1-2 15:4", datetime)
	}
	if err != nil {
		t, err = time.ParseInLocation("2006-1-2", datetime, time.Local)
	}
	return t
}
