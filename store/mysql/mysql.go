package mysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"database/sql"
	"github.com/pkritiotis/outbox/store"
)

type Message struct {
	store.Message
}

type Settings struct {
	MySQLUsername string
	MySQLPass     string
	MySQLHost     string
	MySQLPort     string
	MySQLDB       string
}

type Store struct {
	db *sql.DB
}

func NewStore(settings Settings) (*Store, error) {
	db, err := sql.Open("mysql",
		fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True",
			settings.MySQLUsername, settings.MySQLPass, settings.MySQLHost, settings.MySQLPort, settings.MySQLDB))

	if err != nil || db.Ping() != nil {
		log.Fatalf("failed to connect to database %v", err)
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s Store) ClearLocksWithDurationBeforeDate(duration time.Duration, time time.Time) error {
	panic("implement me")
}

func (s Store) UpdateMessageLockByState(lockID string, lockedOn time.Time, state store.MessageState) error {
	panic("implement me")
}

func (s Store) UpdateMessageByID(message store.Message) error {
	panic("implement me")
}

func (s Store) ClearLocksByLockID(lockID string) error {
	panic("implement me")
}

func (s Store) UpdateMessage(message store.Message) error {
	panic("implement me")
}

func (s Store) GetMessagesByLockID(lockID string) ([]store.Message, error) {
	rows, err := s.db.Query(fmt.Sprintf(`SELECT * from outbox WHERE lock_id = %v AND locked_by = NULL AND locked_on = NULL AND processed_on = NULl`, lockID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// An album slice to hold data from returned rows.
	var messages []store.Message

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var msg store.Message
		if err := rows.Scan(&msg.ID, &msg.Key, &msg.Headers, &msg.Body,
			&msg.Topic); err != nil {
			return messages, err
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return messages, err
	}
	return messages, nil
}

func (s Store) SaveTx(message store.Message, tx *sql.Tx) error {
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
