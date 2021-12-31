package configurable

import (
	"log"
	"os"
)

type Logger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func createLogger() *logger {
	l := &logger{l: log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)}
	return l
}

var _ Logger = (*logger)(nil)

type logger struct {
	l *log.Logger
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.output("DEBUG", format, v...)

}
func (l *logger) Infof(format string, v ...interface{}) {
	l.output("INFO", format, v...)

}
func (l *logger) Warnf(format string, v ...interface{}) {
	l.output("WARN", format, v...)

}
func (l *logger) Errorf(format string, v ...interface{}) {
	l.output("ERROR", format, v...)
}

func (l *logger) output(level, format string, v ...interface{}) {
	if len(v) == 0 {
		l.l.Print(level + " " + format)
		return
	}
	l.l.Printf(level+" "+format, v...)
}
