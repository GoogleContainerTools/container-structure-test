package types

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type LogLevel struct {
	Level log.Level
}

func (llevel *LogLevel) String() string {
	return llevel.Level.String()
}

func (llevel *LogLevel) Set(s string) error {
	level, err := log.ParseLevel(s)
	if err != nil {
		return fmt.Errorf(`%s is not a valid LogLevel.
Please Provide one of %s"`, s, log.AllLevels)
	}
	llevel.Level = level
	return nil
}

func (llevel *LogLevel) Type() string {
	return "types.LogLevel"
}

func NewLogLevel(val string, p *LogLevel) *LogLevel {
	level, err := log.ParseLevel(val)
	if err != nil {
		return nil
	}
	*p = LogLevel{
		Level: level,
	}
	return p
}
