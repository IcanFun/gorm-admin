package gorm_admin

type FilterOperator string
type DatabaseType string

const (
	FilterOperatorLike           FilterOperator = "like"
	FilterOperatorIn             FilterOperator = "in"
	FilterOperatorBetween        FilterOperator = "between"
	FilterOperatorGreater        FilterOperator = ">"
	FilterOperatorGreaterOrEqual FilterOperator = ">="
	FilterOperatorEqual          FilterOperator = "="
	FilterOperatorNotEqual       FilterOperator = "!="
	FilterOperatorLess           FilterOperator = "<"
	FilterOperatorLessOrEqual    FilterOperator = "<="
)

const (
	Int    DatabaseType = "int"
	Float  DatabaseType = "float"
	String DatabaseType = "string"
)
