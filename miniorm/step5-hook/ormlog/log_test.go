package ormlog

import (
	"testing"
	"time"
)

func TestSetLevel(t *testing.T) {
	SetLevel(ErrorLevel)
	Debugf("print log @ %s", time.Now())
	Infof("print log @ %s", time.Now())
	Warnf("print log @ %s", time.Now())
	Errorf("print log @ %s", time.Now())
}
