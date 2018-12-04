package slowlog

import (
	"strings"

	"github.com/gomodule/redigo/redis"
)

type SlowLogData struct {
	Id            int64
	Timestamp     int64
	Duration      int
	Cmd           string
	Key           string
	Args          []string
	ClientAddress string
	ClientName    string
}

type slowLog struct {
	Conn redis.Conn
}

func NewSlowLog(conn redis.Conn) *slowLog {
	return &slowLog{
		Conn: conn,
	}
}

func (s *slowLog) FetchSlowLog() ([]SlowLogData, error) {
	results, err := redis.Values(s.Conn.Do("SLOWLOG", "GET"))
	var slowLogArr []SlowLogData
	for _, item := range results {
		entry, err := redis.Values(item, nil)
		if err != nil {
			//	fmt.Println("Error loading slowlog values: %v", err)
			continue
		}
		var log SlowLogData
		var args []string
		redis.Scan(entry,
			&log.Id,
			&log.Timestamp,
			&log.Duration,
			&args,
			&log.ClientAddress,
			&log.ClientName)
		// This splits up the args into cmd, key, args.
		argsLen := len(args)
		if argsLen > 0 {
			log.Cmd = strings.Replace(args[0], "\n", "", -1)
		}
		if argsLen > 1 {
			log.Key = strings.Replace(args[1], "\n", "", -1)
		}
		slowLogArr = append(slowLogArr, log)
	}
	return slowLogArr, err

}

func (s *slowLog) Ping() bool {
	result, err := redis.String(s.Conn.Do("PING"))
	if err != nil {
		return false
	}
	if result == "PONG" {
		return true
	}
	return false
}
