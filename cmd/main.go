package main

import (
	trace "go_demo/pkg/trace"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	go trace.StartTrace([]trace.TraceConfig{
		{
			Address: "77xRsWrEJsvX5pNRCD5zYq8pJt5XnZ9xYwK6gT4599gv",
			Remark:  "77x-最近经常割看第二笔有没有机会",
		},
		{
			Address: "6Te6gGSsJLa8Gwo5k5kZeJjA5hfuceU5e2FKcvnPoxPG",
			Remark:  "6Te-1201新发现",
		},
	})
	// go stream.Subscribe()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	println("Exiting...")
}
