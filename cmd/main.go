package main

import (
	stream "go_demo/pkg/stream"
	"os"
	"os/signal"
	"syscall"
	// trace "go_demo/pkg/trace"
)

func main() {
	go stream.RunStream()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	println("Exiting...")
}
