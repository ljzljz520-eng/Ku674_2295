package observability

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

type Logger struct {
	base *slog.Logger
	mu   sync.Mutex
}

func NewLogger() *Logger {
	return &Logger{base: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))}
}

func (l *Logger) Info(message string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.base.Info(message, args...)
}
func (l *Logger) Error(message string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.base.Error(message, args...)
}
func (l *Logger) BatchSummary(batchID string, accepted, pending int) {
	l.Info(fmt.Sprintf("batch %s summary", batchID), "accepted", accepted, "pending", pending)
}
