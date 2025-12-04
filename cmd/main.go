package main

import (
	"os"
	"os/signal"
	"syscall"

	stream "go_demo/pkg/stream"
)

func main() {
	go stream.SubscribeFilteredTransactions()

	// 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	println("Exiting...")
}
