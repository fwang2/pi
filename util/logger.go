package util

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func str2level(loglevel string) log.Level {
	switch loglevel {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "fatal":
		return log.FatalLevel
	case "trace":
		return log.TraceLevel
	default:
		fmt.Printf("Unknow log level %s\n", loglevel)
		os.Exit(1)
	}

	return log.InfoLevel

}

func NewLogger() *log.Logger {
	var log = log.New()

	// check environment
	loglevel, ok := os.LookupEnv("PI_LOG")

	if !ok {
		log.SetLevel(str2level("info"))
	} else {
		log.SetLevel(str2level(loglevel))
	}
	return log
}
