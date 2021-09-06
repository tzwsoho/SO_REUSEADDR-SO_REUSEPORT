//go:build linux || freebsd || dragonfly || darwin
// +build linux freebsd dragonfly darwin

package main

import (
	"log"

	"golang.org/x/sys/unix"
)

func SetFdOpt(fd uintptr) {
	if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		log.Fatal(err)
		return
	}

	if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		log.Fatal(err)
		return
	}
}
