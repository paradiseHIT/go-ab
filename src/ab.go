/*
 * @Author: xing paradisehit@gmil.com
 * @Date: 2022-12-03 12:16:57
 * @LastEditors: xing paradisehit@gmil.com
 * @LastEditTime: 2022-12-04 23:41:54
 * @FilePath: /go-ab/ab.go
 * @Description:
 */
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"
)

type Request struct {
	content []byte
}

const maxIdleConn = 500

// max request body size is set 16M
const maxRequestBodySize = 16 * 1024 * 1024
const (
	headerRegexp = `^([\w-]+):\s*(.+)`
)

var mutex sync.Mutex

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func (c *AbBenchmark) Ab() {
	var wg sync.WaitGroup
	wg.Add(c.thread_num)
	tr := &http.Transport{
		DisableKeepAlives:   c.disable_keepalive,
		MaxIdleConnsPerHost: min(c.thread_num, maxIdleConn),
	}
	//reuse the http client to get high performance
	client := &http.Client{Transport: tr, Timeout: time.Duration(c.time_out) * time.Second}

	request_per_thread := int(c.request_num / c.thread_num)
	if request_per_thread <= 0 {
		glog.Errorf("request per thread is %d\n", request_per_thread)
		return
	}
	for i := 0; i < c.thread_num; i++ {
		glog.V(3).Infof("Treading %d start", i)
		go func(thread_id int) {
			c.run(client, thread_id, request_per_thread)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (c *AbBenchmark) VerifyConfig() {
	if c.request_file == "" && c.method == "POST" {
		usageAndExit("Please specify request_file when http method set POST")
	}
	if c.url == "" {
		usageAndExit("")
	}
	if c.duration > 0 {
		c.request_num = math.MaxInt32
	}
}

func (c *AbBenchmark) PrintConfig() {
	glog.Infof("config thread_num:\t\t\t%d", c.thread_num)
	glog.Infof("config request_num:\t\t\t%d", c.request_num)
	glog.Infof("config request_file:\t\t\t%s", c.request_file)
	glog.Infof("config request_is_path:\t\t\t%v", c.request_is_path)
	glog.Infof("config url:\t\t\t\t%s", c.url)
	if c.QPS > 0 {
		glog.Infof("config QPS:\t\t\t\t%d", c.QPS)
	} else {
		glog.Infof("config QPS:\t\t\t\tunlimit")
	}
	glog.Infof("config method:\t\t\t\t%s", c.method)
	glog.Infof("config request_num_in_file:\t\t%d", c.request_num_in_file)
	glog.Infof("config content_type:\t\t\t%s", c.content_type)
	glog.Infof("config time_out:\t\t\t%d", c.time_out)
	glog.Infof("config duration:\t\t\t%v", c.duration)
	if len(c.hs) > 0 {
		glog.Infof("Header:")
		for _, h := range c.hs {
			glog.Infof("\t%s", h)
		}
	}
	if c.debug {
		glog.Infof("config mode:\t\t\tdebug")
	}

	fmt.Println()
}

func (c *AbBenchmark) Init() {
	c.stop_sig = make(chan struct{}, c.thread_num)
	//LoadRequestsFromFile should be in front of PrintConfig for config request_num_in_file
	if c.request_file != "" {
		c.LoadRequestsFromFile()
	}
	c.InitRequest()
}
func (c *AbBenchmark) InitRequest() {
	var err error
	/*
	   Don't New Request every time in MakeRequest, it will cost large amount of connections
	   We can clone request from one, just set different body
	*/
	c.req_glob, err = http.NewRequest(c.method, c.url, nil)
	if err != nil {
		usageAndExit(err.Error())
	}
	// set content-type
	header := make(http.Header)
	header.Set("Content-Type", c.content_type)
	c.req_glob.Header = header

	// set any other additional repeatable headers
	for _, h := range c.hs {
		match, err := parseInputWithRegexp(h, headerRegexp)
		if err != nil {
			usageAndExit(err.Error())
		}
		header.Set(match[1], match[2])
	}

}

func (c *AbBenchmark) run(client *http.Client, thread_id int, request_per_thread int) {
	var throttle <-chan time.Time
	if c.QPS > 0 {
		throttle = time.Tick(time.Duration(1e6/(float32(c.QPS)/float32(c.thread_num))) * time.Microsecond)
	}

	//TODO to add exit signal when user press ctrl+C
	for j := 0; j < request_per_thread; j++ {
		select {
		case <-c.stop_sig:
			return
		default:
			if c.QPS > 0 {
				<-throttle
			}
			c.MakeRequest(client, thread_id, j)
		}
	}
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request, body []byte) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	if len(body) > 0 {
		r2.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	r2.ContentLength = int64(len(body))
	return r2
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (c *AbBenchmark) MakeRequest(client *http.Client, thread_id int, request_index int) int {
	var body []byte
	if c.request_num_in_file > 0 {
		body = c.GetRequest(rand.Intn(c.request_num_in_file)).content
	}

	req := cloneRequest(c.req_glob, body)

	s_time := now()
	resp, err := client.Do(req)
	if err == nil {
		if c.debug {
			resp_data, resp_err := ioutil.ReadAll(resp.Body)
			glog.Infof("return code :%d", resp.StatusCode)
			if resp_err == nil {
				glog.Infof("return resp :%s", resp_data)
			} else {
				glog.Infof("return error :%s", resp_err.Error())
			}
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	} else {
		glog.Errorf("post %s error, %s", body, err.Error())
		panic("post error")
	}
	t := (int)((now() - s_time) / 1000)
	c.AppendResult(t)
	return t
}

func (c *AbBenchmark) Report(pcts *[]int) {
	sort.Ints(c.result_arr)
	pcts_value := c.ArrayInfo(pcts)
	fmt.Printf("Thread num\t\t\t\t%d\n", c.thread_num)
	fmt.Printf("Total request:\t\t\t\t%d\n", len(c.result_arr))
	for i := 0; i < len(*pcts); i++ {
		fmt.Printf("%d%%\t\t\t\t\t%dus\n", (*pcts)[i], pcts_value[i])
	}
	var total_time = 0
	for i := 0; i < len(c.result_arr); i++ {
		total_time += (c.result_arr)[i]
	}
	fmt.Printf("Avg\t\t\t\t\t%dus\n", total_time/len(c.result_arr))
	if c.QPS > 0 {
		fmt.Printf("QPS(set)\t\t\t\t%d\n", c.QPS)
	} else {
		total_time_ms := float64(total_time) / 1000.0
		fmt.Printf("QPS(real)\t\t\t\t%.2f\n", float64(len(c.result_arr)*c.thread_num)*1000/total_time_ms)
		if c.debug {
			glog.Infof("result cnt:%d\ttotal cost cpu time:%.2fms\tthread num:%d\n", len(c.result_arr), total_time_ms, c.thread_num)
		}
	}

}
func (c *AbBenchmark) AppendResult(req int) {
	mutex.Lock()
	c.result_arr = append(c.result_arr, req)
	mutex.Unlock()
}

func (c *AbBenchmark) ArrayInfo(pcts *[]int) []int {
	total_len := len(c.result_arr)
	var pcts_value []int
	for i := 0; i < len(*pcts); i++ {
		offset := max(0, int(float64(total_len)*float64((*pcts)[i])/100.0)-1)
		pcts_value = append(pcts_value, c.result_arr[offset])
	}
	return pcts_value
}

func (c *AbBenchmark) LoadRequestsFromFile() {
	rfile, err := os.Open(c.request_file)
	if err != nil {
		glog.Errorf("Open %s failed\n", c.request_file)
		panic(err.Error())
	}
	buf := []byte{}
	fileScanner := bufio.NewScanner(rfile)
	fileScanner.Buffer(buf, maxRequestBodySize)
	for fileScanner.Scan() {
		var r Request
		read_line := fileScanner.Text()
		if c.request_is_path {
			f, err := os.Open(read_line)
			check(err)
			body := make([]byte, maxRequestBodySize)
			n, err := f.Read(body)
			check(err)
			r.content = body[:n]
			check(f.Close())
		} else {
			r.content = []byte(read_line)
		}
		c.request_que = append(c.request_que, r)
	}
	c.request_num_in_file = len(c.request_que)
	glog.V(3).Infof("request_num_in_file:%d", len(c.request_que))
}

func (c *AbBenchmark) GetRequest(index int) Request {
	return c.request_que[index]
}

func (c *AbBenchmark) Stop() {
	for i := 0; i < c.thread_num; i++ {
		c.stop_sig <- struct{}{}
	}
}

func parseInputWithRegexp(input, regx string) ([]string, error) {
	re := regexp.MustCompile(regx)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 1 {
		return nil, fmt.Errorf("could not parse the provided input; input = %v", input)
	}
	return matches, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
