package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	LOG_DIR = "logs/"
	CWD_KEY = "cwd"
)

type hook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (h *hook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range h.Writer {
		if _, err = w.Write([]byte(line)); err != nil {
			return err
		}
	}
	return nil
}

// Levels define on which log levels this hook will trigger
func (h *hook) Levels() []logrus.Level {
	return h.LogLevels
}

var entry *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func GetLogger() *Logger {
	return &Logger{entry}
}

func (l *Logger) GetLoggerWithField(k string, v interface{}) Logger {
	return Logger{l.WithField(k, v)}
}

// init initialize logrus for logging
func init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
		},
		DisableColors: false,
		FullTimestamp: true,
	}

	var cwd string
	if value, ok := os.LookupEnv(CWD_KEY); ok {
		cwd = filepath.Join(value, LOG_DIR)
	} else {
		cwd = filepath.Join(".", LOG_DIR)
	}

	if err := os.MkdirAll(cwd, 0755); err != nil || os.IsExist(err) {
		panic(fmt.Errorf("Can't create dir '%s' %w", cwd, err))
	}

	logFilename := time.Now().Format("2006-01-02") + ".log"
	logFile, err := os.OpenFile(filepath.Join(cwd, logFilename), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(fmt.Errorf("Failed to open file '%s' %w", logFilename, err))
	}

	// Send all logs to nowhere by default
	l.SetOutput(io.Discard)
	l.AddHook(&hook{
		Writer:    []io.Writer{logFile, os.Stdout},
		LogLevels: logrus.AllLevels,
	})

	l.SetLevel(logrus.TraceLevel)
	entry = logrus.NewEntry(l)
}
