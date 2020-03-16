package log

import (
	"testing"
)

func TestLogger_Log(t *testing.T) {
	logger := &Logger{Level: DebugLevel}

	logger.Infof("a %v", 1)
}
