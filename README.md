# Description

Nifty tool which collect the SLOWLOG from redis server and forward to SumoLogic

This tool use Go modules and is compiled with Go 1.11.X

## Usage
Just run `forwarder --help` to get the latest help

```bash
usage: forwarder --redis-server=REDIS-SERVER --redis-password=REDIS-PASSWORD --sumologic-url=SUMOLOGIC-URL [<flags>]

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --redis-server=REDIS-SERVER    Redis server address
  --redis-password=REDIS-PASSWORD
                                 Password for Redis
  --query-interval=10s           Redis SlowLog interval Query
  --sumologic-url=SUMOLOGIC-URL  SumoLogic Collector URL as give by SumoLogic
  --sumologic-source-category=""
                                 Override default Source Category
  --sumologic-source-name=""     Override default Source Name
  --sumologic-source-host=""     Override default Source Host
  --version                      Show application version.

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
  [] Add Buffer to avoid to much pressure
  [] Put retrieve/sending into GoRoutine
  [] More test coverage
  [] Dedup. before sending SumoLogic
