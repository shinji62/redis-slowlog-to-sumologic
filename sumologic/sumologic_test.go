package sumologic

import (
	"testing"
	"time"

	"github.com/shinji62/redis-slowlog-to-sumologic/slowlog"
	"github.com/stretchr/testify/assert"
)

var (
	slog = &slowlog.SlowLogData{
		Id:            3829,
		Timestamp:     1543565632,
		Duration:      25100,
		Cmd:           "evalsha",
		Key:           "2e231c7b52db341bdeb6601fcc19c631630a0112",
		Args:          []string{},
		ClientAddress: "18.182.111.165:24926"}
)

func TestFormatEvents(t *testing.T) {
	sumo := NewSumoLogic("http://127.0.0.1", "hostname", "sumologicName", "sumologicCategory", "0.0.0", 100*time.Millisecond)
	b := sumo.FormatEvents(*slog)
	assert.Equal(t, b, "{\"Id\":3829,\"Timestamp\":1543565632,\"Duration\":25100,\"Cmd\":\"evalsha\",\"Key\":\"2e231c7b52db341bdeb6601fcc19c631630a0112\",\"Args\":[],\"ClientAddress\":\"18.182.111.165:24926\",\"ClientName\":\"\"}")
}
