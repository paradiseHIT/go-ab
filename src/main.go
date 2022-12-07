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
	"net/http"
	"runtime"

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
	flag.BoolVar(&(ab.disable_keepalive), "disable_keepalive", false, "Disable keep-alive, prevents re-use of TCP, connections between different HTTP requests.")
	flag.IntVar(&(ab.cpus), "cpus", runtime.GOMAXPROCS(-1), "Number of used cpu cores.")

}

func main() {
	flag.Parse()
	defer glog.Flush()

	runtime.GOMAXPROCS(ab.cpus)
	ab.VerifyConfig()
	ab.PrintConfig()
	ab.LoadRequestsFromFile()
	ab.InitRequest()
	ab.Ab()

	pcts := []int{50, 60, 70, 80, 90, 99}
	ab.Report(&pcts)
}