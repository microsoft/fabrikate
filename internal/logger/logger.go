package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// lock is a global mutex lock to gain control of logrus.<SetLevel|SetOutput>
var lock = sync.Mutex{}

// SetLevelDebug sets the standard logger level to Debug
func SetLevelDebug() {
	lock.Lock()
	logrus.SetLevel(logrus.DebugLevel)
	lock.Unlock()
}

// SetLevelInfo sets the standard logger level to Info
func SetLevelInfo() {
	lock.Lock()
	logrus.SetLevel(logrus.InfoLevel)
	lock.Unlock()
}

// Trace logs a message at level Trace to stdout.
func Trace(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Trace(args...)
	lock.Unlock()
}

// Tracef logs a message at level Trace to stdout.
func Tracef(format string, args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Tracef(format, args...)
	lock.Unlock()
}

// Traceln logs a message at level Trace to stdout.
func Traceln(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Traceln(args...)
	lock.Unlock()
}

// Debug logs a message at level Debug to stdout.
func Debug(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Debug(args...)
	lock.Unlock()
}

// Debugf logs a message at level Debug to stdout.
func Debugf(format string, args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Debugf(format, args...)
	lock.Unlock()
}

// Debugln logs a message at level Debug to stdout.
func Debugln(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Debugln(args...)
	lock.Unlock()
}

// Info logs a message at level Info to stdout.
func Info(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Info(args...)
	lock.Unlock()
}

// Infof logs a message at level Info to stdout.
func Infof(format string, args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Infof(format, args...)
	lock.Unlock()
}

// Infoln logs a message at level Info to stdout.
func Infoln(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Infoln(args...)
	lock.Unlock()
}

// Warn logs a message at level Warn to stdout.
func Warn(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Warn(args...)
	lock.Unlock()
}

// Warnf logs a message at level Warn to stdout.
func Warnf(format string, args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Warnf(format, args...)
	lock.Unlock()
}

// Warnln logs a message at level Warn to stdout.
func Warnln(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Warnln(args...)
	lock.Unlock()
}

// Error logs a message at level Error to stderr.
func Error(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stderr)
	logrus.Error(args...)
	lock.Unlock()
}

// Errorf logs a message at level Error to stdout.
func Errorf(format string, args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Errorf(format, args...)
	lock.Unlock()
}

// Errorln logs a message at level Error to stdout.
func Errorln(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Errorln(args...)
	lock.Unlock()
}

// Fatal logs a message at level Fatal to stderr then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stderr)
	logrus.Fatal(args...)
	lock.Unlock()
}

// Fatalf logs a message at level Fatal to stdout.
func Fatalf(format string, args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Fatalf(format, args...)
	lock.Unlock()
}

// Fatalln logs a message at level Fatal to stdout.
func Fatalln(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Fatalln(args...)
	lock.Unlock()
}

// Panic logs a message at level Panic to stderr; calls panic() after logging.
func Panic(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stderr)
	logrus.Panic(args...)
	lock.Unlock()
}

// Panicf logs a message at level Panic to stdout.
func Panicf(format string, args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Panicf(format, args...)
	lock.Unlock()
}

// Panicln logs a message at level Panic to stdout.
func Panicln(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Panicln(args...)
	lock.Unlock()
}

func init() {
	// Setup logger defaults
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = "02-01-2006 15:04:05"
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)
	logrus.SetOutput(os.Stdout) // Set output to stdout; set to stderr by default
	logrus.SetLevel(logrus.InfoLevel)
}
