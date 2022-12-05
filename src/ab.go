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
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/golang/glog"
)

type Request struct {
	content string
}

var request_que []Request
var mutex sync.Mutex

func ab(request_file string, requestURL string, thread_num int, total_request_num int) {
	request_num_in_file := BuildRequests(request_file)
	request_cnt_per_thread := int(total_request_num / thread_num)
	remain_cnt := int(total_request_num % thread_num)
	var request_by_threads []int
	var request_time_arr []int

	for i := 0; i < thread_num; i++ {
		request_by_threads = append(request_by_threads, request_cnt_per_thread)
	}
	for i := 0; i < remain_cnt; i++ {
		request_by_threads[i]++
	}
	for i := 0; i < thread_num; i++ {
		glog.Infof("thread %d : %d", i, request_by_threads[i])
	}

	var wg sync.WaitGroup
	wg.Add(thread_num)

	for i := 0; i < thread_num; i++ {
		glog.Infof("Treading %d start", i)
		go func(thread_id int) {
			rand.Seed(time.Now().Unix())
			for j := 0; j < request_by_threads[thread_id]; j++ {
				r := getRequest(rand.Intn(request_num_in_file))
				req := []byte(r.content)
				//us
				start_time := time.Now().Nanosecond() / 1000
				resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(req))
				t := time.Now().Nanosecond()/1000 - start_time
				if err != nil {
					glog.Errorf("post %s error, %s", r.content, err.Error())
					panic("post error")
				}
				respBytes, err := ioutil.ReadAll(resp.Body)
				AppendRequest(&request_time_arr, t)
				glog.Infof("thread %d, request %d, randInt %d, value:%s, result:%s", thread_id, j, rand.Intn(request_num_in_file), r.content, string(respBytes))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	glog.Infof("len(request_time_arr):%d, total_request_num:%d", len(request_time_arr), total_request_num)
	for i := 0; i < len(request_time_arr); i++ {
		glog.Infof("request_cost[%d]=%d", i, request_time_arr[i])
	}
	sort.Ints(request_time_arr)
	fmt.Println(request_time_arr)
	pcts := []int{50, 90}
	pcts_value := ArrayInfo(&request_time_arr, &pcts)
	fmt.Println(pcts_value)
}
func AppendRequest(arr_in *[]int, req int) {
	mutex.Lock()
	*arr_in = append(*arr_in, req)
	mutex.Unlock()
}

func ArrayInfo(arr_in *[]int, pcts *[]int) []int {
	total_len := len(*arr_in)
	var pcts_value []int
	glog.Infof("len(*arr_in):%d", len(*arr_in))
	glog.Infof("len(*pcts):%d", len(*pcts))
	for i := 0; i < len(*pcts); i++ {
		offset := int(total_len*(*pcts)[i]/100) - 1
		glog.Infof("offset:%d", offset)
		pcts_value = append(pcts_value, (*arr_in)[offset])
	}
	return pcts_value
}

func BuildRequests(request_file string) int {
	rfile, err := os.Open(request_file)
	if err != nil {
		glog.Errorf("Open %s failed\n", request_file)
		return -1
	}
	fileScanner := bufio.NewScanner(rfile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		var r Request
		r.content = fileScanner.Text()
		request_que = append(request_que, r)
	}
	return len(request_que)

}

func getRequest(index int) Request {
	return request_que[index]
}
