package model

const (
	TransactionLogTypeUpdate = 1
	TransactionLogTypeInsert = 2
)

type TransactionLog struct {
	Id           int64  `db:"id"`
	Tid          string `db:"tid"`
	Type         int    `db:"type"`
	BeforeCol    string `db:"before_col"`
	AfterCol     string `db:"after_col"`
	Table        string `db:"table_name"`
	Connection   string `db:"connection_str"`
	PrimaryKey   string `db:"primary_key"`
	PrimaryValue string `db:"primary_value"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}
