// Package mysql provides a mysql implementation of the outbox.Store interface
package mysql

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // needed for loading mysql driver
	"github.com/pkritiotis/outbox"
)

// Settings contain the mysql settings
type Settings struct {
	MySQLUsername string
	MySQLPass     string
	MySQLHost     string
	MySQLPort     string
	MySQLDB       string
}

// Store implements a mysql Store
type Store struct {
	db *sql.DB
}

// NewStore constructor
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

// ClearLocksWithDurationBeforeDate clears all records with the provided id
func (s Store) ClearLocksWithDurationBeforeDate(time time.Time) error {
	_, err := s.db.Exec(
		`UPDATE outbox 
		SET
			locked_by=NULL,
			locked_on=NULL
		WHERE locked_on < ?
		`,
		time,
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateRecordLockByState updated the lock information based on the state
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

// UpdateRecordByID updates the provided record based on its id
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
			processed_on=?,
		    number_of_attempts=?,
		    last_attempted_on=?,
		    error=?
		WHERE id = ?
		`,
		msgData.Bytes(),
		rec.State,
		rec.CreatedOn,
		rec.LockID,
		rec.LockedOn,
		rec.ProcessedOn,
		rec.NumberOfAttempts,
		rec.LastAttemptOn,
		rec.Error,
		rec.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

// ClearLocksByLockID clears lock information of the records with the provided id
func (s Store) ClearLocksByLockID(lockID string) error {
	_, err := s.db.Exec(
		`UPDATE outbox 
		SET 
			locked_by=NULL,
			locked_on=NULL
		WHERE locked_by = ?
		`,
		lockID)
	if err != nil {
		return err
	}
	return nil
}

// GetRecordsByLockID returns the records of the provided id
func (s Store) GetRecordsByLockID(lockID string) ([]outbox.Record, error) {
	rows, err := s.db.Query(
		"SELECT id, data, state, created_on,locked_by,locked_on,processed_on,number_of_attempts,last_attempted_on,error from outbox WHERE locked_by = ?",
		lockID,
	)
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
		scanErr := rows.Scan(&rec.ID, &data, &rec.State, &rec.CreatedOn, &rec.LockID, &rec.LockedOn, &rec.ProcessedOn, &rec.NumberOfAttempts, &rec.LastAttemptOn, &rec.Error)
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

// AddRecordTx stores the record in the db within the provided transaction tx
func (s Store) AddRecordTx(rec outbox.Record, tx *sql.Tx) error {
	msgBuf := new(bytes.Buffer)
	msgEnc := gob.NewEncoder(msgBuf)
	encErr := msgEnc.Encode(rec.Message)

	if encErr != nil {
		return encErr
	}
	q := "INSERT INTO outbox (id, data, state, created_on,locked_by,locked_on,processed_on,number_of_attempts,last_attempted_on,error) VALUES (?,?,?,?,?,?,?,?,?,?)"

	_, err := tx.Exec(q,
		rec.ID,
		msgBuf.Bytes(),
		rec.State,
		rec.CreatedOn,
		rec.LockID,
		rec.LockedOn,
		rec.ProcessedOn,
		rec.NumberOfAttempts,
		rec.LastAttemptOn,
		rec.Error)
	if err != nil {
		return err
	}
	return nil
}

// RemoveRecordsBeforeDatetime removes records before the provided datetime
func (s Store) RemoveRecordsBeforeDatetime(expiryTime time.Time) error {
	_, err := s.db.Exec(
		`DELETE FROM outbox 
		WHERE created_on < ?
		`,
		expiryTime)
	if err != nil {
		return err
	}
	return nil
}
