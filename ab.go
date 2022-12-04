/*
 * @Author: xing paradisehit@gmil.com
 * @Date: 2022-12-03 12:16:57
 * @LastEditors: xing paradisehit@gmil.com
 * @LastEditTime: 2022-12-04 01:25:35
 * @FilePath: /go-ab/ab.go
 * @Description:
 */
package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang/glog"
)

type Request struct {
	content string
}

var request_que []Request

func ab(request_file string, requestURL string, thread_num int, total_request_num int) {
	request_num_in_file := buildRequests(request_file)
	request_cnt_per_thread := int(total_request_num / thread_num)
	remain_cnt := int(total_request_num % thread_num)
	var request_by_threads []int
	var request_time_queue Queue

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
				request_time_queue.Push(t)
				glog.Infof("thread %d, request %d, randInt %d, value:%s, result:%s", thread_id, j, rand.Intn(request_num_in_file), r.content, string(respBytes))
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	glog.Infof("len(request_time_arr):%d, total_request_num:%d", request_time_queue.Len(), total_request_num)
	tmp := 0
	for i := request_time_queue.List.Front(); i != nil; i = i.Next() {
		glog.Infof("request_cost[%d]=%d", tmp, i.Value)
		tmp++
	}
}

func buildRequests(request_file string) int {
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
