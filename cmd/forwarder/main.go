package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shinji62/redis-slowlog-to-sumologic/logging"
	"github.com/shinji62/redis-slowlog-to-sumologic/slowlog"
	"github.com/shinji62/redis-slowlog-to-sumologic/sumologic"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	alias            = kingpin.Flag("env-alias", "Environment alias use for Prometheus metrics(qa,prod,...)").Envar("ENV_ALIAS").Required().String()
	rAddr            = kingpin.Flag("redis.server", "Redis server address").Required().Envar("REDIS_SERVER").String()
	rPassword        = kingpin.Flag("redis.password", "Password for Redis").Required().Envar("REDIS_PASSWORD").String()
	rsizeSlowLog     = kingpin.Flag("redis.slowlog", "Numbers of SlowLog to fetch (default 100)").Envar("REDIS_SLOWLOG").Default("100").Int()
	qInterval        = kingpin.Flag("query-interval", "Redis SlowLog interval Query").Envar("SUMOLOGIC_QUERY_INT").Default("10s").Duration()
	dupClearInterval = kingpin.Flag("dups-cache-ttl", "Interval which duplicate cache is cleared").Envar("SUMOLOGIC_QUERY_INT").Default("60s").Duration()
	sURL             = kingpin.Flag("sumologic.url", "SumoLogic Collector URL as give by SumoLogic").Required().Envar("SUMOLOGIC_URL").String()
	sSourceCategory  = kingpin.Flag("sumologic.source.category", "Override default Source Category").Envar("SUMOLOGIC_CAT").Default("").String()
	sSourceName      = kingpin.Flag("sumologic.source.name", "Override default Source Name").Default("").String()
	sSourceHost      = kingpin.Flag("sumologic.source.host", "Override default Source Host").Default("").String()
	listenAddress    = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9121").String()
	metricPath       = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
)

const (
	ExitCodeOk    = 0
	ExitCodeError = 1 + iota
)

var (
	version     = "0.0.0"
	builddate   = ""
	commit_sha1 = ""
)

func main() {
	//logging init
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	kingpin.Version(version)
	kingpin.Parse()

	//Connect to redis server
	rConnection, err := redis.Dial("tcp", *rAddr,
		redis.DialConnectTimeout(2*time.Second),
		redis.DialReadTimeout(15*time.Second),
		redis.DialWriteTimeout(3*time.Second),
		redis.DialPassword(*rPassword))

	if err != nil {
		fmt.Println(err)
		os.Exit(ExitCodeError)
	}

	r := slowlog.NewSlowLog(rConnection, *dupClearInterval)
	sClient := sumologic.NewSumoLogic(
		*sURL,
		*sSourceHost,
		*sSourceName,
		*sSourceCategory,
		version,
		2*time.Second)

	//Adding Prometheus metrics
	//Build info
	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "redis_slowlog_exporter_build_info",
		Help: "redis_slowlog_exporter_build_info",
	}, []string{"version", "commit_sha", "build_date", "golang_version"})
	buildInfo.WithLabelValues(version, commit_sha1, builddate, runtime.Version()).Set(1)

	//Number of SlowLog Processed since application start
	slowLogProcessed := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "redis_slowlog_exporter_total_slowlog",
		Help: "Total of SlowLog processed since application start",
		ConstLabels: prometheus.Labels{"redis_server": *rAddr,
			"alias": *alias},
	})
	//Number of Error SlowLog Processed since application start
	slowLogErrorProcessed := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "redis_error_slowlog_exporter_total_slowlog",
		Help: "Total of SlowLog processed since application start",
		ConstLabels: prometheus.Labels{"redis_server": *rAddr,
			"alias": *alias},
	})

	prometheus.MustRegister(buildInfo)
	prometheus.MustRegister(slowLogProcessed)
	prometheus.MustRegister(slowLogErrorProcessed)
	http.Handle(*metricPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Redis SlowLog Forwarder to Sumologic</title></head>
             <body>
             <h1>Redis SlowLog Forwarder to Sumologic</h1>
             <p><a href='` + *metricPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	go func() {
		logging.Error.Fatal(http.ListenAndServe(*listenAddress, nil))
	}()

	q := make(chan os.Signal)
	signal.Notify(q, os.Interrupt)
	ctx, cancelf := context.WithCancel(context.Background())

	slowLogsResults := make(chan []slowlog.SlowLogData)
	// get slowLogs and send to SumoLogic forwarder
	go func(c chan []slowlog.SlowLogData, ctx context.Context) {
		//Ticker for X sec we query the SlowLog from Redis
		ticker := time.NewTicker(*qInterval)
		for {
			select {
			case <-ticker.C:

				slowLogResult, err := r.FetchSlowLog(*rsizeSlowLog)
				if err != nil {
					slowLogErrorProcessed.Inc()
					logging.Error.Println(err)
					continue
				}
				c <- slowLogResult
			case <-ctx.Done():
				fmt.Print(ctx.Err())
				return
			}
		}
	}(slowLogsResults, ctx)

	// receive logs from channel and send it via the forwarder
	go func(c <-chan []slowlog.SlowLogData, ctx context.Context) {
		for {
			select {
			case sLogs := <-c:
				for _, sLog := range sLogs {
					formated := sClient.FormatEvents(sLog)
					slowLogProcessed.Inc()
					go sClient.SendLogs(formated)
				}
			case <-ctx.Done():
				fmt.Print(ctx.Err())
				return
			}
		}
	}(slowLogsResults, ctx)

	<-q
	close(slowLogsResults)
	cancelf()
	fmt.Print("Forwarder quitting...")
}
