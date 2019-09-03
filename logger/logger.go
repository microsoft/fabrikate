package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// lock is a global mutex lock to gain control of logrus.<SetLevel|SetOutput>
var lock = sync.RWMutex{}

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

// Debug logs a message at level Debug to stdout.
func Debug(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Debug(args...)
	lock.Unlock()
}

// Info logs a message at level Info to stdout.
func Info(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Info(args...)
	lock.Unlock()
}

// Warn logs a message at level Warn to stdout.
func Warn(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stdout)
	logrus.Warn(args...)
	lock.Unlock()
}

// Error logs a message at level Error to stderr.
func Error(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stderr)
	logrus.Error(args...)
	lock.Unlock()
}

// Fatal logs a message at level Fatal to stderr then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stderr)
	logrus.Fatal(args...)
	lock.Unlock()
}

// Panic logs a message at level Panic to stderr; calls panic() after logging.
func Panic(args ...interface{}) {
	lock.Lock()
	logrus.SetOutput(os.Stderr)
	logrus.Panic(args...)
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
