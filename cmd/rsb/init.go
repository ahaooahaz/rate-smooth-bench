package main

import "runtime"

func init() {
	initEnv()
}

func initEnv() (err error) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	return
}
