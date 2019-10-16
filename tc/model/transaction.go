package model

import "strconv"

const (
	TransactionStatusNormal = 0
	TransactionStatusCommitted = 1
	TransactionStatusRollbacking = 2
	TransactionStatusRollbacked = 3
)

type Transaction struct {
	Id int64 `db:"id"`
	Appid string `db:"appid"`
	Name string	`db:"name"`
	Status int	`db:"status"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (t *Transaction) ToStr() string {
	idStr := strconv.FormatInt(t.Id,10)
	statusStr := strconv.Itoa(t.Status)
	return "{"+idStr+","+t.Appid+","+t.Name+","+statusStr+","+t.CreatedAt+","+t.UpdatedAt+"}"
}