package logger

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
)

var ffile *os.File

func Info(msg string) {
	log.Info(msg)
}

func WARN(msg string) {
	log.Warning(msg)
}

func ERROR(msg string) {
	log.Error(msg)
	os.Stdout.WriteString(fmt.Sprintf("%+v\n", errors.New("")))
	ffile.WriteString(fmt.Sprintf("%+v\n", errors.New("")))
}
