package sumologic

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/shinji62/redis-slowlog-to-sumologic/logging"
	"github.com/shinji62/redis-slowlog-to-sumologic/slowlog"
)

type SumoLogicEvents struct {
}

type SumoLogic struct {
	httpClient           http.Client
	forwarderVersion     string
	sumoCategory         string
	sumoName             string
	sumoHost             string
	sumoURL              string
	sumoPostMinimumDelay time.Duration
	timerBetweenPost     time.Time
	customMetadata       string
}

func NewSumoLogic(url string, host string, name string, category string, expVersion string, connectionTimeOut time.Duration) *SumoLogic {
	return &SumoLogic{
		forwarderVersion: expVersion,
		sumoHost:         host,
		sumoName:         name,
		sumoURL:          url,
		sumoCategory:     category,
		httpClient:       http.Client{Timeout: connectionTimeOut},
	}
}

//FormatEvents
//Format SlowLog Interface to flat string
func (s *SumoLogic) FormatEvents(slowLog slowlog.SlowLogData) string {
	var msg []byte
	sumologicMessage, err := json.Marshal(slowLog)
	if err == nil {
		msg = sumologicMessage
	}
	return string(msg)

}

func (s *SumoLogic) SendLogs(logStringToSend string) {
	logging.Trace.Println("Attempting to send to Sumo Endpoint: " + s.sumoURL)
	if logStringToSend != "" {
		var buf bytes.Buffer
		g := gzip.NewWriter(&buf)
		g.Write([]byte(logStringToSend))
		g.Close()
		request, err := http.NewRequest("POST", s.sumoURL, &buf)
		if err != nil {
			logging.Error.Printf("http.NewRequest() error: %v\n", err)
			return
		}
		request.Header.Add("Content-Encoding", "gzip")
		request.Header.Add("X-Sumo-Client", "redis-forwarder v"+s.forwarderVersion)

		if s.sumoName != "" {
			request.Header.Add("X-Sumo-Name", s.sumoName)
		}
		if s.sumoHost != "" {
			request.Header.Add("X-Sumo-Host", s.sumoHost)
		}
		if s.sumoCategory != "" {
			request.Header.Add("X-Sumo-Category", s.sumoCategory)
		}
		//checking the timer before first POST intent
		for time.Since(s.timerBetweenPost) < s.sumoPostMinimumDelay {
			logging.Trace.Println("Delaying Post because minimum post timer not expired")
			time.Sleep(100 * time.Millisecond)
		}
		response, err := s.httpClient.Do(request)

		if (err != nil) || (response.StatusCode != 200 && response.StatusCode != 302 && response.StatusCode < 500) {
			logging.Info.Println("Endpoint dropped the post send")
			logging.Info.Println("Waiting for 300 ms to retry")
			time.Sleep(300 * time.Millisecond)
			statusCode := 0
			err := Retry(func(attempt int) (bool, error) {
				var errRetry error
				request, err := http.NewRequest("POST", s.sumoURL, &buf)
				if err != nil {
					logging.Error.Printf("http.NewRequest() error: %v\n", err)
				}
				request.Header.Add("Content-Encoding", "gzip")
				request.Header.Add("X-Sumo-Client", "redis-forwarder v"+s.forwarderVersion)

				if s.sumoName != "" {
					request.Header.Add("X-Sumo-Name", s.sumoName)
				}
				if s.sumoHost != "" {
					request.Header.Add("X-Sumo-Host", s.sumoHost)
				}
				if s.sumoCategory != "" {
					request.Header.Add("X-Sumo-Category", s.sumoCategory)
				}
				//checking the timer before POST (retry intent)
				for time.Since(s.timerBetweenPost) < s.sumoPostMinimumDelay {
					logging.Trace.Println("Delaying Post because minimum post timer not expired")
					time.Sleep(100 * time.Millisecond)
				}
				response, errRetry = s.httpClient.Do(request)

				if errRetry != nil {
					logging.Error.Printf("http.Do() error: %v\n", errRetry)
					logging.Info.Println("Waiting for 300 ms to retry after error")
					time.Sleep(300 * time.Millisecond)
					return attempt < 5, errRetry
				} else if response.StatusCode != 200 && response.StatusCode != 302 && response.StatusCode < 500 {
					logging.Info.Println("Endpoint dropped the post send again")
					logging.Info.Println("Waiting for 300 ms to retry after a retry ...")
					statusCode = response.StatusCode
					time.Sleep(300 * time.Millisecond)
					return attempt < 5, errRetry
				} else if response.StatusCode == 200 {
					logging.Trace.Println("Post of logs successful after retry...")
					s.timerBetweenPost = time.Now()
					statusCode = response.StatusCode
					return true, err
				}
				return attempt < 5, errRetry
			})
			if err != nil {
				logging.Error.Println("Error, Not able to post after retry")
				logging.Error.Printf("http.Do() error: %v\n", err)
				return
			} else if statusCode != 200 {
				logging.Error.Printf("Not able to post after retry, with status code: %d", statusCode)
			}
		} else if response.StatusCode == 200 {
			logging.Trace.Println("Post of logs successful")
			s.timerBetweenPost = time.Now()
		}

		if response != nil {
			defer response.Body.Close()
		}
	}
}

//------------------Retry Logic Code-------------------------------

// MaxRetries is the maximum number of retries before bailing.
var MaxRetries = 10
var errMaxRetriesReached = errors.New("exceeded retry limit")

// Func represents functions that can be retried.
type Func func(attempt int) (retry bool, err error)

// Do keeps trying the function until the second argument
// returns false, or no error is returned.
func Retry(fn Func) error {
	var err error
	var cont bool
	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if attempt > MaxRetries {
			return errMaxRetriesReached
		}
	}
	return err
}

// IsMaxRetries checks whether the error is due to hitting the
// maximum number of retries or not.
func IsMaxRetries(err error) bool {
	return err == errMaxRetriesReached
}
