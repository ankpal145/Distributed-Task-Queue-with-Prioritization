package logger

import (
	"context"
	"os"

	"github.com/ankurpal/distributed-task-queue/config"
	"github.com/ankurpal/distributed-task-queue/pkg/contexts"
	"github.com/sirupsen/logrus"
)

const (
	FieldServiceName   = "service_name"
	FieldRequestID     = "request_id"
	FieldCorrelationID = "correlation_id"
	FieldUserID        = "user_id"
	FieldTaskID        = "task_id"
	FieldWorkerID      = "worker_id"
	FieldURLPath       = "url_path"
	FieldMethod        = "method"
)

type Logger struct {
	logger *logrus.Logger
}

func SetupLogger(cfg *config.Configuration) (*Logger, error) {
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		return nil, err
	}

	log := logrus.New()
	log.Out = os.Stdout
	log.Level = level

	if cfg.Logging.Format == "json" {
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		}
	} else {
		log.Formatter = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		}
	}

	log.WithField(FieldServiceName, cfg.Name)

	logger := &Logger{logger: log}
	return logger, nil
}

func (l *Logger) GetLogger() *logrus.Logger {
	return l.logger
}

func (l *Logger) InfoWithCtx(ctx context.Context, format string, args ...interface{}) {
	l.withCtx(ctx).Infof(format, args...)
}

func (l *Logger) WarnWithCtx(ctx context.Context, format string, args ...interface{}) {
	l.withCtx(ctx).Warnf(format, args...)
}

func (l *Logger) ErrorWithCtx(ctx context.Context, format string, args ...interface{}) {
	l.withCtx(ctx).Errorf(format, args...)
}

func (l *Logger) DebugWithCtx(ctx context.Context, format string, args ...interface{}) {
	l.withCtx(ctx).Debugf(format, args...)
}

func (l *Logger) FatalWithCtx(ctx context.Context, format string, args ...interface{}) {
	l.withCtx(ctx).Fatalf(format, args...)
}

func (l *Logger) withCtx(ctx context.Context) *logrus.Entry {
	defer func() {
		if r := recover(); r != nil {
			l.logger.WithFields(logrus.Fields{
				"panic": r,
			}).Error("Recovered in logger context extraction")
		}
	}()

	entry := l.logger.WithFields(logrus.Fields{})

	reqContext := contexts.GetRequestContext(ctx)
	if reqContext != nil {
		if reqContext.RequestID != "" {
			entry = entry.WithField(FieldRequestID, reqContext.RequestID)
		}
		if reqContext.CorrelationID != "" {
			entry = entry.WithField(FieldCorrelationID, reqContext.CorrelationID)
		}
		if reqContext.UserID != "" {
			entry = entry.WithField(FieldUserID, reqContext.UserID)
		}
		if reqContext.URL != "" {
			entry = entry.WithField(FieldURLPath, reqContext.URL)
		}
		if reqContext.Method != "" {
			entry = entry.WithField(FieldMethod, reqContext.Method)
		}
	}

	if taskID := ctx.Value("task_id"); taskID != nil {
		entry = entry.WithField(FieldTaskID, taskID)
	}

	if workerID := ctx.Value("worker_id"); workerID != nil {
		entry = entry.WithField(FieldWorkerID, workerID)
	}

	return entry
}

func (l *Logger) Get(ctx context.Context) *logrus.Entry {
	return l.withCtx(ctx)
}

func (l *Logger) Info(msg string, fields ...logrus.Fields) {
	entry := l.logger.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Info(msg)
}

func (l *Logger) Error(msg string, err error, fields ...logrus.Fields) {
	entry := l.logger.WithFields(logrus.Fields{})
	if err != nil {
		entry = entry.WithError(err)
	}
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Error(msg)
}

func (l *Logger) Warn(msg string, fields ...logrus.Fields) {
	entry := l.logger.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Warn(msg)
}

func (l *Logger) Debug(msg string, fields ...logrus.Fields) {
	entry := l.logger.WithFields(logrus.Fields{})
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Debug(msg)
}
