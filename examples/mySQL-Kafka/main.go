package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/pkritiotis/outbox"
	"github.com/pkritiotis/outbox/broker/kafka"
	"github.com/pkritiotis/outbox/logs/zerolog"
	"github.com/pkritiotis/outbox/store/mysql"
	"os"
	"time"
)

type SampleMessage struct {
	message string
}

var (
	errChan  chan error
	doneChan chan struct{}

	sqlSettings mysql.Settings
	brokerAddr  string
)

func init() {
	errChan = make(chan error)
	doneChan = make(chan struct{})
	sqlSettings = mysql.Settings{
		MySQLUsername: "root",
		MySQLPass:     "a123456",
		MySQLHost:     "localhost",
		MySQLDB:       "outbox",
		MySQLPort:     "3306",
	}
	brokerAddr = "localhost:29092"
}

func main() {

	defer func() { doneChan <- struct{}{} }()

	// Initialize the logger
	logger := zerolog.NewZerologAdapter("go-outbox", os.Stdout)

	//Initialize the sql store
	store, err := mysql.NewStore(sqlSettings, logger)
	if err != nil {
		fmt.Printf("Could not initialize the store: %v", err)
		os.Exit(1)
	}

	//Initialize the message broker
	c := sarama.NewConfig()
	c.Producer.Return.Successes = true
	broker, err := kafka.NewBroker([]string{brokerAddr}, c)
	if err != nil {
		fmt.Printf("Could not initialize the message broker: %v", err)
		os.Exit(1)
	}

	//Initialize and run the dispatcher
	settings := outbox.DispatcherSettings{
		ProcessInterval:           20 * time.Second,
		LockCheckerInterval:       600 * time.Minute,
		CleanupWorkerInterval:     60 * time.Second,
		MaxLockTimeDuration:       5 * time.Minute,
		MessagesRetentionDuration: 1 * time.Minute,
	}
	dispatcher := outbox.NewDispatcher(store, logger, broker, settings, "1")
	dispatcher.Run(errChan, doneChan)

	go func() {
		err = <-errChan
		fmt.Printf(err.Error())
	}()

	//Initialize the outbox service
	publisher := outbox.NewPublisher(store)

	//Open a db connection and perform a transaction
	db, _ := openDbConnection()
	tx, _ := db.BeginTx(context.Background(), nil)

	encodedData, _ := json.Marshal(SampleMessage{message: "ok"})
	publisher.Send(outbox.Message{
		Key:     "sampleKey",
		Headers: nil,
		Body:    encodedData,
		Topic:   "sampleTopic",
	}, tx)
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Could not commit the sql transaction: %v", err)
		os.Exit(1)
	}
	<-doneChan
}

func openDbConnection() (*sql.DB, error) {
	return sql.Open("mysql",
		fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True",
			sqlSettings.MySQLUsername, sqlSettings.MySQLPass, sqlSettings.MySQLHost, sqlSettings.MySQLPort, sqlSettings.MySQLDB))
}
