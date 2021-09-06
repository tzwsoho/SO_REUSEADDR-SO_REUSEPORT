package main

import (
	"log"

	"golang.org/x/sys/windows"
)

func SetFdOpt(fd uintptr) {
	if err := windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
		return
	}
}
