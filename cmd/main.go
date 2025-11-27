package main

import (
	trace "go_demo/pkg/trace"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	go trace.StartTrace("5MhSK3DYj41ubC3n2YeRjhrXxyQQmdXNKtFoYep2mg3x")
	// go stream.Subscribe()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	println("Exiting...")
}
