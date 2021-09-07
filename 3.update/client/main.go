package main

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s IP", os.Args[0])
	}

	// Method 1
	// 每次都使用同一个连接
	// for {
	// 	t := strconv.FormatInt(time.Now().UnixNano(), 10)
	// 	for retries := 0; retries < 5; retries++ {
	// 		res, err := http.Post("http://"+os.Args[1]+":1234/ping", "text/plain", bytes.NewReader([]byte(t)))
	// 		if err != nil {
	// 			pc, file, line, _ := runtime.Caller(0)
	// 			log.Fatalln(pc, file, line, t, err)
	// 		} else if http.StatusOK != res.StatusCode {
	// 			pc, file, line, _ := runtime.Caller(0)
	// 			log.Fatalln(pc, file, line, t, res)
	// 		} else {
	// 			b, err := ioutil.ReadAll(res.Body)
	// 			if err != nil {
	// 				pc, file, line, _ := runtime.Caller(0)
	// 				log.Fatalln(pc, file, line, t, err)
	// 			} else {
	// 				log.Printf("%d Recv: %s, org: %s", time.Now().UnixNano(), string(b), t)
	// 				break
	// 			}
	// 		}
	// 	}
	// }

	// Method 2
	// 每次使用不同的连接
	for {
		client := http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 20 * time.Second,
		}

		t := strconv.FormatInt(time.Now().UnixNano(), 10)
		for retries := 0; retries < 5; retries++ {
			req, err := http.NewRequest("POST", "http://"+os.Args[1]+":1234/ping", bytes.NewReader([]byte(t)))
			if err != nil {
				pc, file, line, _ := runtime.Caller(0)
				log.Fatalln(pc, file, line, t, err)
			} else {
				req.Close = true
				req.Header.Add("Connection", "Close")

				if res, err := client.Do(req); nil != err { // CentOS 7 测试：旧服务器关闭，或新服务器开启，都有可能产生错误
					pc, file, line, _ := runtime.Caller(0)
					log.Fatalln(pc, file, line, t, err)
				} else if http.StatusOK != res.StatusCode {
					pc, file, line, _ := runtime.Caller(0)
					log.Fatalln(pc, file, line, t, res)
				} else {
					_, err := ioutil.ReadAll(res.Body)
					if err != nil {
						pc, file, line, _ := runtime.Caller(0)
						log.Fatalln(pc, file, line, t, err)
					} else {
						// log.Printf("%d Recv: %s, org: %s", time.Now().UnixNano(), string(b), t)
						break
					}
				}
			}
		}
	}
}
