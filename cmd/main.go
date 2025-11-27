package main

import (
	// trace "go_demo/pkg/trace"
	stream "go_demo/pkg/stream"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// go trace.GetTransaction("3Em8p9NpiWijcc2c9sGy35eLo7Wgg9x7fLhLy5WeD51JtFqj2r74djVyaaT7H21q94xJPaxERu35NZ6nzCrkdUJp")
	go stream.Subscribe()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	println("Exiting...")
}
