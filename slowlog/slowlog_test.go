package slowlog

import (
	"fmt"
	"testing"

	"github.com/rafaeljusto/redigomock"
)

var (
	slowlogResult = []interface{}{
		[]interface{}{
			int64(950),
			int64(1483706756),
			int64(144),
			[]interface{}{
				[]uint8("SLOWLOG"),
				[]uint8("GET"),
				[]uint8("128"),
			},
		},
		[]interface{}{
			int64(949),
			int64(1483706756),
			int64(13),
			[]interface{}{
				[]uint8("CONFIG"),
				[]uint8("GET"),
				[]uint8("slowlog-max-len"),
			},
		},
		[]interface{}{
			int64(948),
			int64(1483706756),
			int64(60),
			[]interface{}{
				[]uint8("INFO"),
			},
		},
		[]interface{}{
			int64(947),
			int64(1483706717),
			int64(121),
			[]interface{}{
				[]uint8("SLOWLOG"),
				[]uint8("GET"),
				[]uint8("128"),
			},
		},
		[]interface{}{
			int64(946),
			int64(1483706717),
			int64(15),
			[]interface{}{
				[]uint8("CONFIG"),
				[]uint8("GET"),
				[]uint8("slowlog-max-len"),
			},
		},
	}
)

func TestNewRedis(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("PING").Expect(string("PONG"))

	sl := NewSlowLog(conn)
	if ping := sl.Ping(); ping != true {
		t.Fatalf("Command (%v) return (%v) ", cmd.Name, ping)
	}
}

func TestFetchSlowLog(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("SLOWLOG", "GET").Expect(slowlogResult)
	sl := NewSlowLog(conn)
	result, err := sl.FetchSlowLog()
	fmt.Println(conn.Stats(cmd))
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 5 {
		t.Fatalf("Expected SlOWLOG to return %v result", 5)
	}
}
