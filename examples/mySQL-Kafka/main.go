package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
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
	c := sarama.NewConfig()
	c.Producer.Return.Successes = true
	broker := kafka.NewBroker([]string{"localhost:9092"}, c)
	settings := outbox.DispatcherSettings{
		ProcessIntervalSeconds:     20,
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
	s := outbox.NewDispatcher(store, broker, settings, "1")
	errChan := make(chan error)
	doneChan := make(chan struct{})
	s.Run(errChan, doneChan)
	defer func() { doneChan <- struct{}{} }()
	err := <-errChan
	fmt.Printf(err.Error())
}
