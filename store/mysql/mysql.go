package mysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/pkritiotis/go-outbox/outbox"
	"time"
)
import sql "database/sql"

type Message struct {
	outbox.Message
}

type MySQL struct {
	db *sql.DB
}

func (s MySQL) UpdateMessageLockByState(lockID string, lockedOn time.Time, state outbox.MessageState) error {
	panic("implement me")
}

func (s MySQL) UpdateMessageByID(message outbox.Message) error {
	panic("implement me")
}

func (s MySQL) ClearLocksBeforeDate(time time.Time) error {
	panic("implement me")
}

func (s MySQL) ClearLocksByLockID(lockID string) error {
	panic("implement me")
}

func (s MySQL) UpdateMessages(messages []outbox.Message) error {
	panic("implement me")
}

func (s MySQL) UpdateMessage(message outbox.Message) error {
	panic("implement me")
}

func (s MySQL) GetMessagesByLockID(lockID string) ([]outbox.Message, error) {
	rows, err := s.db.Query(fmt.Sprintf(`SELECT * from outbox WHERE lock_id = %v AND locked_by = NULL AND locked_on = NULL AND processed_on = NULl`, lockID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// An album slice to hold data from returned rows.
	var messages []outbox.Message

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var msg outbox.Message
		if err := rows.Scan(&msg.ID, &msg.Key, &msg.Headers, &msg.Body,
			&msg.Topic, &msg.Type); err != nil {
			return messages, err
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return messages, err
	}
	return messages, nil
}

func (s MySQL) SaveTx(message outbox.Message, tx *sql.Tx) error {
	headers := new(bytes.Buffer)
	enc := gob.NewEncoder(headers)
	headerErr := enc.Encode(message.Headers)
	if headerErr != nil {
		return headerErr
	}

	body := new(bytes.Buffer)
	enc = gob.NewEncoder(body)
	bodyErr := enc.Encode(message.Body)
	if bodyErr != nil {
		return bodyErr
	}

	_, err := tx.Exec(fmt.Sprintf(
		`INSERT INTO outbox (id, key, headers, body, topic, type, state, created_on,locked_by,locked_on,processed_on)
			VALUES (%v,'%v','%v','%v',%v,%v,%v,%v,%v,%v)`,
		message.ID,
		message.Key,
		headers,
		body,
		message.Topic,
		message.State,
		message.CreatedOn,
		message.LockID,
		message.LockedOn,
		message.ProcessedOn))
	if err != nil {
		return err
	}
	return nil
}
