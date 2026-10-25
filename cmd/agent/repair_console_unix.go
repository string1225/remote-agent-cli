//go:build !windows

package main

import "os"

func repairConsole() (*os.File, error) { return os.Open("/dev/tty") }
