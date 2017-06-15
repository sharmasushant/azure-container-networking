// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package log

import (
	"fmt"
	"os"
)

// SetTarget sets the log target.
func (logger *Logger) SetTarget(target int) error {
	var err error

	switch target {
	case TargetStderr:
		logger.out = os.Stderr
	case TargetLogfile:
		logger.out, err = os.OpenFile(logger.getLogFileName(), os.O_CREATE|os.O_APPEND|os.O_RDWR, logFilePerm)
	default:
		err = fmt.Errorf("Invalid log target %d", target)
	}

	if err == nil {
		logger.l.SetOutput(logger.out)
		logger.target = target
	}

	return err
}
