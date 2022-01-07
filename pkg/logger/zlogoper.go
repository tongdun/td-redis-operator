package logger

import (
	"database/sql"
)
import _ "github.com/go-sql-driver/mysql"
import "fmt"

type operlog struct {
	Ts       string `json:"ts"`
	Realname string `json:"realname"`
	Resource string `json:"resource"`
	Oper     string `json:"oper"`
	Status   int    `json:"status"`
}

const (
	OperPending = 0
	OperSuccess = 1
	OperFailed  = 2
)

type action interface {
	logOper(realname string, resource string, oper string) int64
	updateStatus(logId int64, status int)
	getOperLogs(resource string) ([]operlog, error)
}

var operlogger action

type emptyOperlogger struct {
}

func LogOper(realname string, resource string, oper string) int64 {
	return operlogger.logOper(realname, resource, oper)
}

func GetOperLogs(resource string) ([]operlog, error) {
	return operlogger.getOperLogs(resource)
}

func UpdateOperStatus(logId int64, status int) {
	operlogger.updateStatus(logId, status)
}

func (lg *emptyOperlogger) logOper(realname string, resource string, oper string) int64 {
	//do nothing....
	return 0
}

func (lg *emptyOperlogger) updateStatus(logId int64, status int) {

}

func (lg *emptyOperlogger) getOperLogs(resouce string) ([]operlog, error) {
	return []operlog{}, nil
}

type mysqlOperlogger struct {
	conn *sql.DB
	db   string
	tab  string
	user string
	pass string
	addr string
}

func (lg *mysqlOperlogger) logOper(realname string, resource string, oper string) int64 {
	sql := fmt.Sprintf("insert into %s(realname,resource,oper) values('%s','%s','%s')", lg.tab, realname, resource, oper)
	rs, err := lg.conn.Exec(sql)
	if err != nil {
		ERROR(err.Error())
		return 0
	}
	logid, _ := rs.LastInsertId()
	return logid
}

func (lg *mysqlOperlogger) updateStatus(logId int64, status int) {
	sql := fmt.Sprintf("update %s set status=%d where id=%d", lg.tab, status, logId)
	_, err := lg.conn.Exec(sql)
	if err != nil {
		ERROR(err.Error())
	}
}

func (lg *mysqlOperlogger) getOperLogs(resouce string) ([]operlog, error) {
	//sql := fmt.Sprintf("select gmt_create,realname,resource,oper,status from %s where resource='%s'", lg.tab, resouce)
	operlogs := []operlog{}
	rows, err := lg.conn.Query("select gmt_create,realname,resource,oper,status from ? where resource = ?",lg.tab, resouce)
	if err != nil {
		return []operlog{}, err
	}
	for rows.Next() {
		var ts, realname, redis_name, oper string
		var status int
		rows.Scan(&ts, &realname, &redis_name, &oper, &status)
		operlogs = append(operlogs, operlog{
			Ts:       ts,
			Realname: realname,
			Resource: redis_name,
			Oper:     oper,
			Status:   status,
		})

	}
	return operlogs, nil
}

func (lg *mysqlOperlogger) dbint() {
	conn_str := fmt.Sprintf("%s:%s@tcp(%s)/%s", lg.user, lg.pass, lg.addr, lg.db)
	conn, err := sql.Open("mysql", conn_str)
	if err != nil {
		ERROR(err.Error())
	}
	lg.conn = conn
}
