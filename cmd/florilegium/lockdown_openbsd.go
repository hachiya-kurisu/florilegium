package main

import (
	"golang.org/x/sys/unix"
)

func Lockdown(path string) {
	unix.PledgePromises("stdio inet rpath wpath cpath")
}
