package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"syscall"
	"time"
)

// 参考：https://segmentfault.com/a/1190000020524323
/*
在 BSD 下：

SO_REUSEADDR       socketA        socketB       Result
---------------------------------------------------------------------
  ON/OFF       192.168.0.1:21   192.168.0.1:21    Error (EADDRINUSE)
  ON/OFF       192.168.0.1:21      10.0.0.1:21    OK
  ON/OFF          10.0.0.1:21   192.168.0.1:21    OK
   OFF             0.0.0.0:21   192.168.1.0:21    Error (EADDRINUSE)
   OFF         192.168.1.0:21       0.0.0.0:21    Error (EADDRINUSE)
   ON              0.0.0.0:21   192.168.1.0:21    OK
   ON          192.168.1.0:21       0.0.0.0:21    OK
  ON/OFF           0.0.0.0:21       0.0.0.0:21    Error (EADDRINUSE)
*/

func main() {
	c := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) (err error) {
			return c.Control(SetFdOpt)
		},
	}

	lsn, err := c.Listen(context.Background(), "tcp4", "0.0.0.0:1234")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Recv From %s: %s", r.RemoteAddr, string(b))
		t := strconv.FormatInt(time.Now().UnixNano(), 10)
		rw.Write([]byte(t))
	})

	err = http.Serve(lsn, nil)
	if err != nil {
		log.Fatal(err)
	}
}
