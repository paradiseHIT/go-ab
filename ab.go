package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/golang/glog"
)

type Request struct {
	content string
}

var request_que []Request

func ab(requestURL string, thread_num int, wg *sync.WaitGroup) {
	for i := 0; i < thread_num; i++ {
		glog.Infof("Treading %d start", i)
		go run(requestURL, i, wg)
	}
}

func buildRequests(request_file string, request_num int) {
	rfile, err := os.Open(request_file)
	if err != nil {
		glog.Errorf("Open %s failed\n", request_file)
		return
	}
	fileScanner := bufio.NewScanner(rfile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		var r Request
		r.content = fileScanner.Text()
		request_que = append(request_que, r)
	}

}

func getRequest(index int) Request {
	return request_que[index]
}

func run(requestURL string, concurrency int) {
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for {
		res, err := http.Post(requestURL, "application/json", bytes.NewBuffer(post_data))
		if done >= request {
			break
		}
	}
	fmt.Printf("Treading %d done\n", i)
}
