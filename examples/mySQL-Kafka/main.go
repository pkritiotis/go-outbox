package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pkritiotis/outbox"
	"github.com/pkritiotis/outbox/broker/kafka"
	"github.com/pkritiotis/outbox/store/mysql"
	"os"
)

type A struct {
	A string
}

func main() {
	sqlSettings := mysql.Settings{
		MySQLUsername: "root",
		MySQLPass:     "my-secret-pw",
		MySQLHost:     "localhost",
		MySQLDB:       "outbox",
		MySQLPort:     "3306",
	}
	store, dbErr := mysql.NewStore(sqlSettings)
	if dbErr != nil {
		os.Exit(1)
	}
	broker := kafka.Broker{}
	settings := outbox.DispatcherSettings{
		MachineID:                  "1",
		ProcessIntervalSeconds:     100,
		LockCheckerIntervalSeconds: 600,
		MaxLockTimeDurationMins:    5,
	}
	repo := outbox.New(store)

	db, _ := sql.Open("mysql",
		fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True",
			sqlSettings.MySQLUsername, sqlSettings.MySQLPass, sqlSettings.MySQLHost, sqlSettings.MySQLPort, sqlSettings.MySQLDB))
	tx, _ := db.BeginTx(context.Background(), nil)

	encodedData, _ := json.Marshal(A{A: "ok"})
	repo.Add(outbox.Message{
		Key:     "sampleKey",
		Headers: nil,
		Body:    encodedData,
		Topic:   "sampleTopic",
	}, tx)
	tx.Commit()
	s := outbox.NewDispatcher(store, broker, settings)
	errChan := make(chan error)
	go s.Run(errChan)
	err := <-errChan
	fmt.Printf(err.Error())
}
