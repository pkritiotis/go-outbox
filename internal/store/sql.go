package store

import (
	"fmt"
	"github.com/pkritiotis/go-outbox/internal/outbox"
)
import sql "database/sql"

type SQL struct {
	db *sql.DB
}

func NewSql(db *sql.DB) *SQL {
	return &SQL{db: db}
}

func (s SQL) SaveTx(message outbox.Message, tx interface{}) error {
	sqlTx := tx.(*sql.Tx)
	_, err := sqlTx.Exec(fmt.Sprintf("INSERT INTO OUTBOX (id) VALUES (", message.ID, ")"))
	if err != nil {
		return err
	}
	// save to db
	// this should be in a transaction
	return nil
}

func (s SQL) Update(message outbox.Message) error {
	panic("implement me")
}

func (s SQL) GetUnprocessed() ([]outbox.Message, error) {
	panic("implement me")
}
