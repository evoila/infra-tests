package main

import (
	"github.com/fatih/color"
	"github.com/op/go-logging"
	"os"
)

var logger = logging.MustGetLogger("core")
var infoBackend = logging.NewLogBackend(os.Stdout, "", 0)
var errBackend = logging.NewLogBackend(os.Stderr, "", 0)

func LogInfo(message string) {
	logging.SetBackend(infoBackend)
	logger.Info(message)
}

func LogInfoF(format string, args ...interface{}) {
	logging.SetBackend(infoBackend)
	logger.Infof(format, args)
}

func LogError(err error, customMessage string) {
	logging.SetBackend(errBackend)
	logger.Errorf(color.RedString("[ERROR] " + customMessage + ": " + err.Error()))
}

func LogErrorF(format string, args ...interface{}) {
	logging.SetBackend(errBackend)
	logger.Errorf(format, args)
}

