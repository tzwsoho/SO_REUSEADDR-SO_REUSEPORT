package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s IP", os.Args[0])
	}

	for i := 0; i < 1; i++ {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: net.ParseIP(os.Args[1]), Port: 1234, Zone: ""})
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			for {
				t := strconv.FormatInt(time.Now().UnixNano(), 10)
				_, err := conn.Write([]byte(t))
				if err != nil {
					log.Printf("Connection %s Write Error: %s", conn.LocalAddr().String(), err.Error())
					return
				}

				var n int
				var buf [0x10000]byte
				n, err = conn.Read(buf[:])
				if err != nil {
					log.Printf("Connection %s Read Error: %s", conn.LocalAddr().String(), err.Error())
					return
				}

				log.Printf("Connection %s Recv: %s", conn.LocalAddr().String(), string(buf[:n]))
			}
		}()
	}

	select {}
}
