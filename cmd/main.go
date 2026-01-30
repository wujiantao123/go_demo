// package main

// import (
// 	get_dev_address "go_demo/pkg/get_dev_address"
// )

// func main() {
// 	get_dev_address.AddCopyAddress()
// 	// go stream.SubscribeFilteredTransactions()
// 	// go checkholding.StartCheckHolding()
// 	// // 等待退出信号
// 	// sig := make(chan os.Signal, 1)
// 	// signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
// 	// <-sig

// 	// println("Exiting...")
// }

package main

import (
	get_dev_address "go_demo/pkg/get_dev_address"
	"log"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		log.Println("run AddCopyAddress")
		get_dev_address.AddCopyAddress()

		<-ticker.C
	}
}
