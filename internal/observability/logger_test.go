package observability

import "testing"

func TestLoggerMethods(t *testing.T) {
	logger := NewLogger()
	logger.Info("test", "key", "value")
	logger.BatchSummary("b", 1, 0)
	logger.Error("test error")
}
