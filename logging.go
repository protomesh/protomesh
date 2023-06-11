package protomesh

type Logger interface {
	Debug(message string, kv ...interface{})
	Info(message string, kv ...interface{})
	Warn(message string, kv ...interface{})
	Error(message string, kv ...interface{})
	Panic(message string, kv ...interface{})
	With(kv ...interface{}) Logger
}
