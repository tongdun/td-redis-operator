package logger

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	. "td-redis-operator/pkg/conf"
)

func init() {
	file, err := os.OpenFile("operator.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	ffile = file
	if err != nil {
		panic(err)
	}
	writers := []io.Writer{
		file,
		os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	log.SetOutput(fileAndStdoutWriter)
	switch Cfg.Logger["loggertype"] {
	case "mysql":
		mysqllogger := &mysqlOperlogger{
			db:   Cfg.Logger["mysqldb"],
			tab:  "cloudlog",
			user: Cfg.Logger["mysqluser"],
			pass: Cfg.Logger["mysqlpass"],
			addr: Cfg.Logger["mysqladdr"],
		}
		mysqllogger.dbint()
		operlogger = mysqllogger
	default:
		operlogger = &emptyOperlogger{}
	}
}
