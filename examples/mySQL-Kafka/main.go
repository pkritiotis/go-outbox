package main

import (
	"github.com/pkritiotis/outbox"
	"github.com/pkritiotis/outbox/broker/kafka"
	"github.com/pkritiotis/outbox/store/mysql"
	"os"
)

func main() {
	store := mysql.Store{}
	broker := kafka.Broker{}
	settings := outbox.DispatcherSettings{
		MachineID:                  "1",
		ProcessIntervalSeconds:     100,
		LockCheckerIntervalSeconds: 600,
		MaxLockTimeDurationMins:    5,
	}
	repo := outbox.NewOutbox(store)

	repo.Add(outbox.Message{
		Key:     "sampleKey",
		Headers: nil,
		Body:    "ok",
		Topic:   "sampleTopic",
	}, nil)

	s := outbox.NewDispatcher(store, broker, settings)
	err := s.Run()
	if err != nil {
		os.Exit(1)
	}
}
