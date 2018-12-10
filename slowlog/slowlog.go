package slowlog

import (
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"
)

const (
	MAX_ARGS = 10
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

func (s *slowLog) FetchSlowLog(size int) ([]SlowLogData, error) {
	var slowLogArr []SlowLogData
	results, err := redis.Values(s.Conn.Do("SLOWLOG", "GET", size))
	if err != nil {
		fmt.Println(err)
		return slowLogArr, err
	}
	for _, item := range results {
		entry, err := redis.Values(item, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var log SlowLogData
		var args []string
		if len(entry) > 5 {
			redis.Scan(entry,
				&log.Id,
				&log.Timestamp,
				&log.Duration,
				&args,
				&log.ClientAddress,
				&log.ClientName)
		} else {
			redis.Scan(entry,
				&log.Id,
				&log.Timestamp,
				&log.Duration,
				nil,
				&args)
		}

		// This splits up the args into cmd, key, args.
		argsLen := len(args)
		if argsLen > 0 {
			log.Cmd = strings.Replace(args[0], "\n", "", -1)
		}
		if argsLen > 1 {
			log.Key = strings.Replace(args[1], "\n", "", -1)

			argsToAdd := 0
			if argsLen > MAX_ARGS {
				argsToAdd = MAX_ARGS
			} else {
				argsToAdd = argsLen
			}
			for a := 1; a < argsToAdd; a++ {
				log.Args = append(log.Args, strings.Replace(args[a], "\n", "", -1))

			}
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
