package model

import "strconv"

const (
	LocalTransactionStatusNormal     = 0
	LocalTransactionStatusCommitted  = 1
	LocalTransactionStatusRollbacked = 2
)

type LocalTransaction struct {
	Id        int64  `db:"id"`
	Tid       string `db:"tid"`
	Appid     string `db:"appid"`
	Name      string `db:"name"`
	Status    int    `db:"status"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (t *LocalTransaction) ToStr() string {
	idStr := strconv.FormatInt(t.Id, 10)
	statusStr := strconv.Itoa(t.Status)
	return "{" + idStr + "," + t.Tid + "," + t.Appid + "," + t.Name + "," + statusStr + "," + t.CreatedAt + "," + t.UpdatedAt + "}"
}
