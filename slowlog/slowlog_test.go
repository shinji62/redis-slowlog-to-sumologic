package slowlog

import (
	"fmt"
	"testing"
	"time"

	"github.com/rafaeljusto/redigomock"
)

var (
	slowlogResult1 = []interface{}{
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
	}

	slowlogResult2 = []interface{}{
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
			int64(952),
			int64(1483706756),
			int64(65),
			[]interface{}{
				[]uint8("INFO"),
			},
		},
		[]interface{}{
			int64(951),
			int64(1483706717),
			int64(30),
			[]interface{}{
				[]uint8("CONFIG"),
				[]uint8("GET"),
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
	}

	slowlogResult3 = []interface{}{
		[]interface{}{
			int64(953),
			int64(1483706756),
			int64(13),
			[]interface{}{
				[]uint8("CONFIG"),
				[]uint8("GET"),
				[]uint8("slowlog-max-len"),
			},
		},
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
			int64(952),
			int64(1483706756),
			int64(65),
			[]interface{}{
				[]uint8("INFO"),
			},
		},
		[]interface{}{
			int64(951),
			int64(1483706717),
			int64(30),
			[]interface{}{
				[]uint8("CONFIG"),
				[]uint8("GET"),
			},
		},
	}
)

func TestNewRedis(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("PING").Expect(string("PONG"))

	sl := NewSlowLog(conn, 60*time.Second)
	if ping := sl.Ping(); ping != true {
		t.Fatalf("Command (%v) return (%v) ", cmd.Name, ping)
	}
}

func TestFetchSlowLog(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("SLOWLOG", "GET", 100).Expect(slowlogResult1)
	sl := NewSlowLog(conn, 60*time.Second)
	result, err := sl.FetchSlowLog(100)
	fmt.Println(conn.Stats(cmd))
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 5 {
		t.Fatalf("Expected SlOWLOG to return %v result", 5)
	}
}
func TestFetchSlowLogMultiple(t *testing.T) {
	conn := redigomock.NewConn()
	cacheReset := 2 * time.Second
	sl := NewSlowLog(conn, cacheReset)

	// make this reusable in this test context
	fn := func(expectedCount int, slowlogResult []interface{}) {
		cmd := conn.Command("SLOWLOG", "GET", 100).Expect(slowlogResult)
		result, err := sl.FetchSlowLog(100)
		fmt.Println(conn.Stats(cmd))
		if err != nil {
			t.Fatal(err)
		}
		if l := len(result); l != expectedCount {
			t.Fatalf("Expected SlOWLOG to return %v result, got %v", expectedCount, l)
		}
	}
	// cache should work
	fn(5, slowlogResult1)
	fn(0, slowlogResult1)
	// cache should have reset
	time.Sleep(cacheReset + (1 * time.Second))
	fn(4, slowlogResult2)
	fn(1, slowlogResult3)
}
