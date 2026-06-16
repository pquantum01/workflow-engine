package config

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewLogger(cfg Config) *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	if strings.ToLower(cfg.LogFormat) == "text" {
		log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	} else {
		log.SetFormatter(&logrus.JSONFormatter{})
	}

	return log
}
