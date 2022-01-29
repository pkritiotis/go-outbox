package mysql

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkritiotis/outbox"
)

type Message struct {
	outbox.Record
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

func (s Store) ClearLocksWithDurationBeforeDate(time time.Time) error {
	_, err := s.db.Exec(
		`UPDATE outbox 
		SET
			locked_by=NULL,
			locked_on=NULL,
		WHERE locked_on < ?
		`,
		time,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s Store) UpdateRecordLockByState(lockID string, lockedOn time.Time, state outbox.RecordState) error {
	_, err := s.db.Exec(
		`UPDATE outbox 
		SET 
			locked_by=?,
			locked_on=?
		WHERE state = ?
		`,
		lockID,
		lockedOn,
		state,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s Store) UpdateRecordByID(rec outbox.Record) error {
	msgData := new(bytes.Buffer)
	enc := gob.NewEncoder(msgData)
	encErr := enc.Encode(rec.Message)
	if encErr != nil {
		return encErr
	}

	_, err := s.db.Exec(
		`UPDATE outbox 
		SET 
			data=?,
			state=?,
			created_on=?,
			locked_by=?,
			locked_on=?,
			processed_on=?
		WHERE id = ?
		`,
		msgData.Bytes(),
		rec.State,
		rec.CreatedOn,
		rec.LockID,
		rec.LockedOn,
		rec.ProcessedOn,
		rec.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s Store) ClearLocksByLockID(lockID string) error {
	_, err := s.db.Exec(
		`UPDATE outbox 
		SET 
			locked_by=NULL,
			locked_on=NULL
		WHERE id = ?
		`,
		lockID)
	if err != nil {
		return err
	}
	return nil
}

func (s Store) GetRecordsByLockID(lockID string) ([]outbox.Record, error) {
	rows, err := s.db.Query("SELECT id, data from outbox WHERE locked_by = ?", lockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// An album slice to hold data from returned rows.
	var messages []outbox.Record

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var rec outbox.Record
		var data []byte
		scanErr := rows.Scan(&rec.ID, &data)
		if scanErr != nil {
			if scanErr == sql.ErrNoRows {
				return messages, nil
			}
			return messages, err
		}

		decErr := gob.NewDecoder(bytes.NewReader(data)).Decode(&rec.Message)
		if decErr != nil {
			return nil, decErr
		}

		messages = append(messages, rec)
	}
	if err = rows.Err(); err != nil {
		return messages, err
	}
	return messages, nil
}

func (s Store) AddRecordTx(rec outbox.Record, tx *sql.Tx) error {
	msgBuf := new(bytes.Buffer)
	msgEnc := gob.NewEncoder(msgBuf)
	encErr := msgEnc.Encode(rec.Message)

	if encErr != nil {
		return encErr
	}
	q := "INSERT INTO outbox (id, data, state, created_on,locked_by,locked_on,processed_on) VALUES (?,?,?,?,?,?,?)"

	_, err := tx.Exec(q,
		rec.ID,
		msgBuf.Bytes(),
		rec.State,
		rec.CreatedOn,
		rec.LockID,
		rec.LockedOn,
		rec.ProcessedOn)
	if err != nil {
		return err
	}
	return nil
}
