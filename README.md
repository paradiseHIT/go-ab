## 简介

为了方便本地压测http服务，编写的支持多线程的ab压测程序

### 编译

```bash
   make clean
   make
```

### 测试

```bash
   sh test.sh
```

### Usage

```text
Usage: go-ab [options...]

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
```