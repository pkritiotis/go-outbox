# Outbox Implementation in Go
[![Go Build & Test](https://github.com/pkritiotis/go-outbox/actions/workflows/build-test.yml/badge.svg)](https://github.com/pkritiotis/go-outbox/actions/workflows/build-test.yml)[![golangci-lint](https://github.com/pkritiotis/go-outbox/actions/workflows/lint.yml/badge.svg)](https://github.com/pkritiotis/go-outbox/actions/workflows/lint.yml)
[![codecov](https://codecov.io/gh/pkritiotis/go-outbox/branch/main/graph/badge.svg?token=KZBBS5MRXP)](https://codecov.io/gh/pkritiotis/go-outbox)

This project provides an implementation of the Transactional Outbox Pattern in Go

# Features
- Send messages within `sql.Tx` transaction through the Outbox Pattern
- Optional Maximum attempts limit for a specific message
- Outbox row locking so that concurrent outbox workers don't process the same records
  - Includes a background worker that cleans record locks after a specified time
- Extensible message broker interface
- Extensible data store interface for sql databases

## Currently supported providers

### Message Brokers
- Kafka

### Database Providers
- MySQL

# Usage
## Create the outbox table
```sql
CREATE TABLE outbox (
        id varchar(100) NOT NULL,
        data BLOB NOT NULL,
        state INT NOT NULL,
        created_on DATETIME NOT NULL,
        locked_by varchar(100) NULL,
        locked_on DATETIME NULL,
        processed_on DATETIME NULL,
        number_of_attempts INT NOT NULL,
        last_attempted_on DATETIME NULL,
        error varchar(1000) NULL
)
```
## Send a message via the outbox
```go

type SampleMessage struct {
	message string
}

func main() {
  
  //Setup the mysql store
	sqlSettings := mysql.Settings{
		MySQLUsername: "root",
		MySQLPass:     "a123456",
		MySQLHost:     "localhost",
		MySQLDB:       "outbox",
		MySQLPort:     "3306",
	}
	store, err := mysql.NewStore(sqlSettings)
	if err != nil {
		fmt.Sprintf("Could not initialize the store: %v", err)
		os.Exit(1)
	}
  
  // Initialize the outbox
	repo := outbox.New(store)

	db, _ := sql.Open("mysql",
		fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True",
			sqlSettings.MySQLUsername, sqlSettings.MySQLPass, sqlSettings.MySQLHost, sqlSettings.MySQLPort, sqlSettings.MySQLDB))

  // Open the transaction
	tx, _ := db.BeginTx(context.Background(), nil)

  // Encode the message in a string format
	encodedData, _ := json.Marshal(SampleMessage{message: "ok"})
  
  // Send the message
	repo.Add(outbox.Message{
		Key:   "sampleKey",
		Body:  encodedData,
		Topic: "sampleTopic",
	}, tx)
  
	tx.Commit()
}

```
## Start the dispatcher
The dispatcher can run on the same or different instance of the applicatin that uses the outbox.
```go
func main() {
  
  //Setup the mysql store
	sqlSettings := mysql.Settings{
		MySQLUsername: "root",
		MySQLPass:     "a123456",
		MySQLHost:     "localhost",
		MySQLDB:       "outbox",
		MySQLPort:     "3306",
	}
	store, err := mysql.NewStore(sqlSettings)
	if err != nil {
		fmt.Sprintf("Could not initialize the store: %v", err)
		os.Exit(1)
	}
  
  //Setup the kafka message broker
	c := sarama.NewConfig()
	c.Producer.Return.Successes = true
	broker, err := kafka.NewBroker([]string{"localhost:29092"}, c)
	if err != nil {
		fmt.Sprintf("Could not initialize the message broker: %v", err)
		os.Exit(1)
	}

  // Initialize the dispatcher
	
	settings := outbox.DispatcherSettings{
		ProcessInterval:     20 * time.Minute,
		LockCheckerInterval: 600 * time.Minute,
		MaxLockTimeDuration: 5 * time.Minute,
	}
  
  d := outbox.NewDispatcher(store, broker, settings, "1")

  // Run the dispatcher
	errChan := make(chan error)
	doneChan := make(chan struct{})
	d.Run(errChan, doneChan)
	defer func() { doneChan <- struct{}{} }()
	err = <-errChan
	fmt.Printf(err.Error())
}

```
