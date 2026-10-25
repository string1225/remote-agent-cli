package main

import "os"

func repairConsole() (*os.File, error) { return os.OpenFile("CONIN$", os.O_RDONLY, 0) }
