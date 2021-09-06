package main

import (
	"context"
	"log"
	"net"
	"syscall"
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

	for {
		conn, err := lsn.Accept()
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("New Connection: %s", conn.RemoteAddr().String())
		}

		go func() {
			for {
				var buf [0x10000]byte
				n, err := conn.Read(buf[:])
				if err != nil {
					log.Printf("Connection %s Read Error: %s", conn.RemoteAddr().String(), err.Error())
					return
				}

				log.Printf("Recv From %s: %+v", conn.RemoteAddr().String(), string(buf[:n]))

				_, err = conn.Write(buf[:n])
				if err != nil {
					log.Printf("Connection %s Write Error: %s", conn.RemoteAddr().String(), err.Error())
					return
				}
			}
		}()
	}
}
