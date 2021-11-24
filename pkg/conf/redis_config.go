package conf

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	RangeRangeType   = "range"
	ListRangeType    = "list"
	IntValueType     = "int"
	StringValueType  = "string"
	DynamicParameter = "dynamic"
)

type RedisConfEntry struct {
	Name           string        `json:"name"`
	Value          interface{}   `json:"value"`
	ValueType      string        `json:"value_type"`
	ValueRangeType string        `json:"value_range_type"`
	ValueRange     []interface{} `json:"value_range"`
	RedisConfType  string        `json:"redis_conf_type"`
	DefaultValue   interface{}   `json:"default_value"`
}

var redisconf []*RedisConfEntry

func NewIntValueType(name string) *RedisConfEntry {
	return &RedisConfEntry{
		Name:          name,
		ValueType:     IntValueType,
		RedisConfType: DynamicParameter,
	}
}

func NewStringValueType(name string) *RedisConfEntry {
	return &RedisConfEntry{
		Name:          name,
		ValueType:     StringValueType,
		RedisConfType: DynamicParameter,
	}
}

func (rc *RedisConfEntry) WithListRange(list []interface{}) *RedisConfEntry {
	rc.ValueRangeType = ListRangeType
	rc.ValueRange = list
	return rc
}

func (rc *RedisConfEntry) WithRangeRange(list []interface{}) *RedisConfEntry {
	rc.ValueRangeType = RangeRangeType
	rc.ValueRange = list
	return rc
}

func (rc *RedisConfEntry) SetDefault(dv interface{}) *RedisConfEntry {
	rc.DefaultValue = dv
	return rc
}

func Getcmmap(cmstr string) map[string]string {
	cmlist := strings.Split(cmstr, "\n")
	cmmap := make(map[string]string)
	for _, cmentry := range cmlist {
		if cmentry == "" {
			continue
		}
		parameter_pair := strings.Split(cmentry, " ")
		key := parameter_pair[0]
		value := strings.Join(parameter_pair[1:], " ")
		cmmap[key] = value
	}
	return cmmap
}

func GetRedisConfList(cmstr string) []RedisConfEntry {
	entries := []RedisConfEntry{}
	cmmap := Getcmmap(cmstr)
	for _, entry := range redisconf {
		cur := *entry
		if cmmap[cur.Name] != "" {
			if cur.ValueType == IntValueType {
				v, _ := strconv.Atoi(cmmap[cur.Name])
				cur.Value = v

			} else {
				cur.Value = cmmap[cur.Name]
			}
		} else {
			cur.Value = cur.DefaultValue
		}
		entries = append(entries, cur)
	}
	return entries
}

func GetRedisCmstr(entries []RedisConfEntry, old_cmstr string) string {
	cmmap := Getcmmap(old_cmstr)
	cmlist := []string{}
	for _, entry := range entries {
		switch entry.Value.(type) {
		case string:
			cmmap[entry.Name] = fmt.Sprintf("%s", entry.Value)
		case int:
			cmmap[entry.Name] = fmt.Sprintf("%d", int(entry.Value.(int)))
		case float32:
			cmmap[entry.Name] = fmt.Sprintf("%d", int(entry.Value.(float32)))
		case float64:
			cmmap[entry.Name] = fmt.Sprintf("%d", int(entry.Value.(float64)))
		}
	}
	for k, v := range cmmap {
		cmlist = append(cmlist, fmt.Sprintf("%s %s", k, v))
	}
	return strings.Join(cmlist, "\n") + "\n"
}

func init() {
	redisconf = []*RedisConfEntry{
		NewStringValueType("activerehashing").WithListRange([]interface{}{"yes", "no"}).SetDefault("yes"),
		NewStringValueType("appendfsync").WithListRange([]interface{}{"everysec", "always", "no"}).SetDefault("everysec"),
		NewStringValueType("appendonly").WithListRange([]interface{}{"yes", "no"}).SetDefault("no"),
		NewIntValueType("hash-max-ziplist-entries").WithRangeRange([]interface{}{32, 5120}).SetDefault(512),
		NewIntValueType("hash-max-ziplist-value").WithRangeRange([]interface{}{32, 640}).SetDefault(64),
		NewIntValueType("hll-sparse-max-bytes").WithRangeRange([]interface{}{0, 15000}).SetDefault(3000),
		NewIntValueType("list-compress-depth").WithRangeRange([]interface{}{0, 3}).SetDefault(0),
		NewStringValueType("maxmemory-policy").WithListRange([]interface{}{"volatile-lru", "allkeys-lru", "volatile-lfu", "allkeys-lfu", "volatile-random", "allkeys-random", "volatile-ttl", "noeviction"}).SetDefault("allkeys-lru"),
		NewIntValueType("maxmemory-samples").WithRangeRange([]interface{}{3, 10}).SetDefault(5),
		NewStringValueType("no-appendfsync-on-rewrite").WithListRange([]interface{}{"yes", "no"}).SetDefault("no"),
		NewIntValueType("set-max-intset-entries").WithRangeRange([]interface{}{32, 5120}).SetDefault(512),
		NewIntValueType("slowlog-log-slower-than").WithRangeRange([]interface{}{100, 10000000}).SetDefault(10000),
		NewIntValueType("slowlog-max-len").WithRangeRange([]interface{}{128, 12800}).SetDefault(128),
		NewStringValueType("stop-writes-on-bgsave-error").WithListRange([]interface{}{"yes", "no"}).SetDefault("no"),
		NewIntValueType("tcp-keepalive").WithRangeRange([]interface{}{0, 256}).SetDefault(256),
		NewIntValueType("timeout").WithRangeRange([]interface{}{0, 10000}).SetDefault(3000),
		NewIntValueType("zset-max-ziplist-entries").WithRangeRange([]interface{}{32, 5120}).SetDefault(128),
		NewIntValueType("zset-max-ziplist-value").WithRangeRange([]interface{}{32, 640}).SetDefault(64),
	}
}
