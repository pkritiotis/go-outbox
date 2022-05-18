# Outbox Implementation in Go
[![Go Build & Test](https://github.com/pkritiotis/go-outbox/actions/workflows/build-test.yml/badge.svg)](https://github.com/pkritiotis/go-outbox/actions/workflows/build-test.yml)[![golangci-lint](https://github.com/pkritiotis/go-outbox/actions/workflows/lint.yml/badge.svg)](https://github.com/pkritiotis/go-outbox/actions/workflows/lint.yml)
[![codecov](https://codecov.io/gh/pkritiotis/go-outbox/branch/main/graph/badge.svg?token=KZBBS5MRXP)](https://codecov.io/gh/pkritiotis/go-outbox)

This project provides an implementation of the Transactional Outbox Pattern in Go

# Features
- Send messages within `sql.Tx` transaction through the Outbox Pattern
- Maximum attempts limit for a specific message
- Outbox row locking so that concurrent outbox workers don't process the same records
  - Includes a background worker that cleans record locks after a specified time
- Extensible message broker interface
- Extensible data store interface for sql databases

## Currently supported providers

### Message Brokers
- Kafka

### Database Providers
- MySQL

# Usage example

# Developer's Handbook

## Design
### Diagram
### Introducing a new message broker
### Introducing a new data strore provider
