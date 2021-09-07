package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
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
	// 初始化监听器
	lc := net.ListenConfig{
		Control: func(network, address string, rc syscall.RawConn) (err error) {
			return rc.Control(SetFdOpt)
		},
	}

	lsn, err := lc.Listen(context.Background(), "tcp4", "0.0.0.0:1234")
	if err != nil {
		pc, file, line, _ := runtime.Caller(0)
		log.Fatal(pc, file, line, err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////

	http.HandleFunc("/ping", func(rw http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			pc, file, line, _ := runtime.Caller(0)
			log.Fatal(pc, file, line, err)
		}

		log.Printf("Recv From %s: %s", r.RemoteAddr, string(b))
		t := strconv.FormatInt(time.Now().UnixNano(), 10)

		rw.Write([]byte(t))
	})

	// 开始处理 HTTP 请求
	httpServer := &http.Server{}
	httpServer.SetKeepAlivesEnabled(false)

	go httpServer.Serve(lsn)

	//////////////////////////////////////////////////////////////////////////////////////////////

	// 删除旧 PID 文件
	const oldPIDFile string = "/tmp/go_update.pid.old"
	if _, err = os.Stat(oldPIDFile); os.IsExist(err) {
		if err = os.Remove(oldPIDFile); nil != err {
			pc, file, line, _ := runtime.Caller(0)
			log.Fatal(pc, file, line, err)
		}
	}

	// 给旧服务发送更新信号
	const pidFile string = "/tmp/go_update.pid"
	if _, errFS := os.Stat(pidFile); nil == errFS { // 找到旧服务 PID 文件
		// 获取旧服务的 PID
		var buf []byte
		if buf, err = ioutil.ReadFile(pidFile); nil != err {
			pc, file, line, _ := runtime.Caller(0)
			log.Fatal(pc, file, line, err)
		}

		var oldPID int64
		if oldPID, err = strconv.ParseInt(string(buf), 10, 16); nil != err {
			pc, file, line, _ := runtime.Caller(0)
			log.Fatal(pc, file, line, err)
		}

		// 找到旧服务进程
		var proc *os.Process
		if proc, err = os.FindProcess(int(oldPID)); nil != err {
			pc, file, line, _ := runtime.Caller(0)
			log.Fatal(pc, file, line, err)
		}

		// 发送信号
		if err = proc.Signal(syscall.Signal(10)); nil != err { // 旧服务进程已结束，有可能是上次运行残留的 PID 文件，直接删掉此文件
			if err = os.Remove(pidFile); nil != err {
				pc, file, line, _ := runtime.Caller(0)
				log.Fatal(pc, file, line, err)
			}
		} else { // 信号发送成功
			// 将旧服务的 PID 文件改名
			if err = os.Rename(pidFile, oldPIDFile); nil != err {
				pc, file, line, _ := runtime.Caller(0)
				log.Fatal(pc, file, line, err)
			}
		}
	} else if !os.IsNotExist(errFS) { // 其他错误直接退出
		pc, file, line, _ := runtime.Caller(0)
		log.Fatal(pc, file, line, errFS)
	}

	// 新建 PID 文件
	err = ioutil.WriteFile(pidFile, []byte(strconv.FormatInt(int64(os.Getpid()), 10)), 0600)
	if nil != err {
		pc, file, line, _ := runtime.Caller(0)
		log.Fatal(pc, file, line, err)
	}

	// 处理服务更新自定义信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.Signal(10), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c

	// 收到新服务更新信号：
	// 1.关闭监听器，不再接受新的连接
	// 2.删除旧 PID 文件，关闭程序

	httpServer.Shutdown(context.Background())
	os.Remove(oldPIDFile)
}
