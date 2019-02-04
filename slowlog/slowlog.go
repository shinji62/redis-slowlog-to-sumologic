package slowlog

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	MAX_ARGS = 10
)

// SlowLogData redis slowlog return data structure
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

type idMap map[int64]bool

type slowLog struct {
	Conn redis.Conn

	processedIDs idMap
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        *sync.Mutex
}

func (m *idMap) clearEvery(ctx context.Context, mtx *sync.Mutex, t time.Duration) {
	// this will clear the duplication map every specified duration
	// this func is blocking, use `go m.clearEvery(...)` instead
	tickTock := time.NewTicker(t)
	for {
		select {
		case <-tickTock.C:
			mtx.Lock()
			(*m) = idMap{}
			mtx.Unlock()
		case <-ctx.Done():
			fmt.Print(ctx.Err())
			return
		}
	}
}

// NewSlowLog create a new slowlog service instance
func NewSlowLog(conn redis.Conn, clearDupsDuration time.Duration) *slowLog {
	sl := &slowLog{
		Conn:         conn,
		processedIDs: idMap{},
		mutex:        &sync.Mutex{},
	}
	sl.ctx, sl.cancel = context.WithCancel(context.Background())
	go sl.processedIDs.clearEvery(sl.ctx, sl.mutex, clearDupsDuration)
	return sl
}

func (s *slowLog) markID(id int64) {
	s.mutex.Lock()
	s.processedIDs[id] = true
	s.mutex.Unlock()
}

func (s *slowLog) isIDExisting(id int64) bool {
	s.mutex.Lock()
	dup, ok := s.processedIDs[id]
	s.mutex.Unlock()
	return dup && ok
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
			_, err = redis.Scan(entry,
				&log.Id,
				&log.Timestamp,
				&log.Duration,
				&args,
				&log.ClientAddress,
				&log.ClientName)
		} else {
			_, err = redis.Scan(entry,
				&log.Id,
				&log.Timestamp,
				&log.Duration,
				&args)
		}
		if err != nil {
			fmt.Println(fmt.Sprintf("Error during redis.Scan: %v", err))
			continue
		}
		if s.isIDExisting(log.Id) {
			fmt.Println(fmt.Sprintf("Skipping SlowLog(id=%v)!", log.Id))
			continue
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
		s.markID(log.Id)
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

// Destroy close all goroutines and destroy the service
func (s *slowLog) Destroy() {
	s.cancel()
	s.Conn.Close()
	s = nil
}
