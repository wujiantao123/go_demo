package main

import (
	trace "go_demo/pkg/trace"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	go trace.StartTrace([]string{"6Te6gGSsJLa8Gwo5k5kZeJjA5hfuceU5e2FKcvnPoxPG"})
	// go stream.Subscribe()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	println("Exiting...")
}
