package pkgCtl

import (
	"log"
	"os"
)

type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Error(v ...any)
}

type defaultLogger struct{ *log.Logger }

func (l defaultLogger) Debug(v ...any) {
	l.Println("[DEBUG]", v)
}

func (l defaultLogger) Info(v ...any) {
	l.Println("[INFO]", v)
}

func (l defaultLogger) Error(v ...any) {
	log.Println()
	l.Println("[ERROR]", v)
}

var DefaultLogger = defaultLogger{
	log.New(os.Stderr, "", log.LstdFlags),
}
