package main

import (
	"golang.org/x/sys/unix"
)

func Lockdown(path string) {
	unix.Unveil(path, "r w")
	unix.UnveilBlock()
	unix.PledgePromises("stdio inet rpath wpath cpath")
}
