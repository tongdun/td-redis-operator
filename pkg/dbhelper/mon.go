package dbhelper

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	. "td-redis-operator/pkg/conf"
	"td-redis-operator/pkg/logger"
	. "td-redis-operator/pkg/redis"
)

var mon *sql.DB

func GetAliveConn(ip_port string, apply_db string, user string, pass string) *sql.DB {
	conn_str := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, ip_port, apply_db)
	conn, err := sql.Open("mysql", conn_str)
	if err != nil {
		logger.WARN(err.Error())
	}
	return conn
}

func GetRedisv1() ([]Redis, error) {
	rediss := []Redis{}
	sql := "select master_name,redis_host,redis_port,region,created_by from redis_sentinels where deleted=0"
	if rows, err := mon.Query(sql); err != nil {
		return rediss, err
	} else {
		for rows.Next() {
			var redis_port int
			var master_name, redis_host, region, created_by string
			rows.Scan(&master_name, &redis_host, &redis_port, &region, &created_by)
			rediss = append(rediss, Redis{
				Name:     fmt.Sprintf("standby-%s", master_name),
				Host:     redis_host,
				Port:     redis_port,
				Dc:       region,
				Env:      "production",
				Realname: created_by,
				Secret:   Cfg.Redissecret,
			})
		}
		defer rows.Close()
	}
	return rediss, nil
}

func ChangeOwnerV1(name string, owner string) error {
	master_name := name[8:]
	sql := fmt.Sprintf("select * from redis_sentinels where master_name='%s'", master_name)
	rows, err := mon.Query(sql)
	if err != nil {
		return err
	}
	if !rows.Next() {
		return errors.New("Not found")
	}
	rows.Close()
	sql = fmt.Sprintf("update redis_sentinels set created_by='%s',updated_by='%s' where master_name='%s'", owner, owner, master_name)
	if _, err := mon.Exec(sql); err != nil {
		return err
	}
	return nil
}

func GetRedisVM() ([]Redis, error) {
	rediss := []Redis{}
	meta := map[string]map[string]string{}
	sql := "select ip,cluster_name,room from machineinfo where app_name='redis'"
	if rows, err := mon.Query(sql); err != nil {
		return rediss, err
	} else {
		for rows.Next() {
			var ip, cluster, room string
			rows.Scan(&ip, &cluster, &room)
			meta[room] = map[string]string{cluster: ip}
		}
		rows.Close()
	}
	sql = fmt.Sprintf("select cluster_name,owner from redis_vm")
	vm_owner := map[string]string{}
	rows, err := mon.Query(sql)
	if err != nil {
		return []Redis{}, err
	}
	for rows.Next() {
		var cluster, owner string
		rows.Scan(&cluster, &owner)
		vm_owner[cluster] = owner
	}
	for dc, css := range meta {
		for cs, ip := range css {
			realname := "administrator"
			if ow := vm_owner[cs]; ow != "" {
				realname = ow
			}
			rediss = append(rediss, Redis{
				Name:     cs,
				Host:     ip,
				Port:     6379,
				Dc:       dc,
				Env:      "production",
				Realname: realname,
				Secret:   Cfg.Redissecret,
			})
		}
	}
	return rediss, nil
}

func FixRedisVMOwner(cluster string, owner string) error {
	sql := fmt.Sprintf("insert into redis_vm(cluster_name,owner) values ('%s','%s') ON DUPLICATE KEY UPDATE owner='%s'", cluster, owner, owner)
	if _, err := mon.Exec(sql); err != nil {
		return err
	}
	return nil
}

func init() {
	mon = GetAliveConn(Cfg.Mon["ip"], Cfg.Mon["db"], Cfg.Mon["user"], Cfg.Mon["password"])
}
