# Description

Nifty tool which collect the SLOWLOG from redis server and forward to SumoLogic

This tool use Go modules and is compiled with Go 1.11.X

## Usage
Just run `forwarder --help` to get the latest help

```bash
usage: forwarder --env-alias=ENV-ALIAS --redis.server=REDIS.SERVER --redis.password=REDIS.PASSWORD --sumologic.url=SUMOLOGIC.URL [<flags>]

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --env-alias=ENV-ALIAS  Environment alias use for Prometheus metrics(qa,prod,...)
      --redis.server=REDIS.SERVER
                             Redis server address
      --redis.password=REDIS.PASSWORD
                             Password for Redis
      --redis.slowlog=100    Numbers of SlowLog to fetch (default 100)
      --query-interval=10s   Redis SlowLog interval Query
      --dups-cache-ttl=60s   Interval which duplicate cache is cleared
      --sumologic.url=SUMOLOGIC.URL
                             SumoLogic Collector URL as give by SumoLogic
      --sumologic.source.category=""
                             Override default Source Category
      --sumologic.source.name=""
                             Override default Source Name
      --sumologic.source.host=""
                             Override default Source Host
      --web.listen-address=":9121"
                             Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"
                             Path under which to expose metrics.
  -l, --log=ERROR... ...     Log levels: -l TRACE -l INFO -l WARNING -l ERROR
      --version              Show application version.
```
## Environment Variable
This application support Environment variable only for

* Redis password `REDIS_PASSWORD`
* Redis server `REDIS_SERVER`


## Build

```
go build cmd/forwarder/main.go '-mod=vendor'
```

## Test
```
go test ./... -mod=vendor -v
```


## Todo

  - [x] Add Buffer to avoid to much pressure  
  - [ ] Put retrieve/sending into GoRoutine  
  - [ ] More test coverage  
  - [x] Dedup. before sending SumoLogic  
