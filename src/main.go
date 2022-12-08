/*
 * @Author: xing paradisehit@gmil.com
 * @Date: 2022-12-03 12:16:57
 * @LastEditors: xing paradisehit@gmil.com
 * @LastEditTime: 2022-12-04 01:24:36
 * @FilePath: /go-ab/main.go
 * @Description:
 */
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/golang/glog"
)

type AbBenchmark struct {
	method            string
	url               string
	thread_num        int
	request_num       int
	QPS               int
	request_file      string
	disable_keepalive bool
	cpus              int
	hs                headerSlice
	content_type      string
	duration          time.Duration

	stop_sig            chan struct{}
	req_glob            *http.Request
	result_arr          []int
	request_num_in_file int
	time_out            int
	request_que         []Request
}

var ab AbBenchmark

func init() {
	flag.IntVar(&(ab.thread_num), "thread_num", 1, "total thread num")
	flag.IntVar(&(ab.request_num), "request_num", 1000, "total request num")
	flag.StringVar(&(ab.request_file), "request_file", "", "request data file, per line for one requst")
	flag.StringVar(&(ab.url), "url", "", "url to accept the request")
	flag.IntVar(&(ab.QPS), "QPS", 0, "rate limit, request per second, default is 0, means use thread_num, no rate limit")
	flag.IntVar(&(ab.time_out), "time_out", 20, "time out per request in seconds")
	flag.StringVar(&(ab.method), "method", "GET", "http method, GET/POST/PUT")
	flag.StringVar(&(ab.content_type), "content_type", "text/html", "Content-type")
	flag.BoolVar(&(ab.disable_keepalive), "disable_keepalive", false, "Disable keep-alive, prevents re-use of TCP, connections between different HTTP requests.")
	flag.IntVar(&(ab.cpus), "cpus", runtime.GOMAXPROCS(-1), "Number of used cpu cores.")
	flag.Var(&(ab.hs), "header", `Custom HTTP header. You can specify as many as needed by repeating the flag.For example, -H "Authorization: ZGI1YWYxNmQw*****" -H "Content-Type: application/xml" .`)
	flag.DurationVar(&(ab.duration), "duration", 0, "Duration of application to send requests. When duration is reached, application stops and exits. If duration is specified, request_num is ignored. Examples: -duration 10s -duration 3m.")
}

var usage = `Usage: go-ab [options...]

Options:
  -thread_num         total thread num, default is 1
  -request_num        total request num, default is 1000
  -request_file       request data file, per line for one requst
  -url                url to accept the request
  -QPS                rate limit, request per second. default is 0, means use
                      thread_num, no rate limit
  -time_out           time out per request in seconds, default is 20
  -method             http method, GET/POST/PUT, default is "GET"
  -content_type       Content-type, default is "text/html"
  -disable_keepalive  Disable keep-alive, prevents re-use of TCP, connections 
                      between different HTTP requests.
  -cpus               Number of used cpu cores.
                      default for current machine %d cores
  -header             Custom HTTP header. You can specify as many as needed by
                      repeating the flag. For example:
                        -H "Authorization: ZGI1YWYxNmQw*****"
                        -H "Content-Type: application/xml"
  -duration           Duration of application to send requests. When duration is
                      reached, application stops and exits. If duration is specified,
                      request_num is ignored. Examples: -duration 10s -duration 3m.

`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage, runtime.NumCPU()))
	}
	flag.Parse()
	defer glog.Flush()

	runtime.GOMAXPROCS(ab.cpus)
	ab.VerifyConfig()

	ab.Init()
	ab.PrintConfig()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		ab.Stop()
	}()
	if ab.duration > 0 {
		go func() {
			time.Sleep(ab.duration)
			ab.Stop()
		}()
	}
	ab.Ab()
	pcts := []int{50, 60, 70, 80, 90, 99}
	ab.Report(&pcts)
}

type headerSlice []string

func (h *headerSlice) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *headerSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}
