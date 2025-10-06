package logger

type Logger interface {
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Printf(string, string, ...any)
	Fatal(...any)
	Close() error
}
