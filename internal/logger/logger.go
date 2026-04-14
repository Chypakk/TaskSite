package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger struct {
	*log.Logger
}

func New() *Logger{
	return &Logger{Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)}
}

type contextKey string

const loggerKey contextKey = "logger"

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey).(*Logger); ok {
		return l
	}
	return New()
}

func RequestID(ctx context.Context) string{
	if reqID, ok := ctx.Value("request_id").(string); ok {
        return reqID
    }
    return "-"
}

func formatKV(args ...any) string {
    if len(args) == 0 {
        return ""
    }
    var sb strings.Builder
    for i := 0; i < len(args); i += 2 {
        if i+1 < len(args) {
            sb.WriteString(fmt.Sprintf(" %s=%v", args[i], args[i+1]))
        }
    }
    return sb.String()
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
    l.Printf("[INFO] [req:%s] %s%s", RequestID(ctx), msg, formatKV(args...))
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
    l.Printf("[WARN] [req:%s] %s%s", RequestID(ctx), msg, formatKV(args...))
}

func (l *Logger) Error(ctx context.Context, msg string, err error, args ...any) {
    l.Printf("[ERROR] [req:%s] %s err=%v%s", RequestID(ctx), msg, err, formatKV(args...))
}