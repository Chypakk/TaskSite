package logger

import (
	"context"
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func New() *Logger{
	return &Logger{Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)}
}

func RequestID(ctx context.Context) string{
	if reqID, ok := ctx.Value("request_id").(string); ok {
        return reqID
    }
    return "-"
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
    l.Printf("[INFO] [req:%s] %s %v", RequestID(ctx), msg, args)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
    l.Printf("[WARN] [req:%s] %s %v", RequestID(ctx), msg, args)
}

func (l *Logger) Error(ctx context.Context, msg string, err error, args ...any) {
    l.Printf("[ERROR] [req:%s] %s err=%v %v", RequestID(ctx), msg, err, args)
}