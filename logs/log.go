package logs

// logs provide an interface that you need to implement to go-outbox logging.
// you can adopt custom logger freely.

// LoggerAdapter - interface that you need to implement to go-outbox logging
type LoggerAdapter interface {
	Trace(msg string, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Fatal(msg string, fields map[string]interface{})
}
