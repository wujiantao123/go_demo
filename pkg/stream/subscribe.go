package stream

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	laserstream "github.com/helius-labs/laserstream-sdk/go"
)

func RunStream() {

	clientConfig := laserstream.LaserstreamConfig{
		Endpoint: "84.32.103.140:10030",
	}

	commitment := laserstream.CommitmentLevel_PROCESSED

	// --- 订阅所有交易 ---
	subReq := &laserstream.SubscribeRequest{
		Transactions: map[string]*laserstream.SubscribeRequestFilterTransactions{
			"all-tx": {}, // 全链交易流
		},
		Commitment: &commitment,
	}

	client := laserstream.NewClient(clientConfig)

	dataCallback := func(update *laserstream.SubscribeUpdate) {
		tx := update.GetTransaction().Transaction
		if tx == nil {
			return
		}
		log.Printf("Received transaction: %s", tx.Signature)

	}

	errorCallback := func(err error) {
		log.Printf("Error: %v", err)
	}

	if err := client.Subscribe(subReq, dataCallback, errorCallback); err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	client.Unsubscribe()
}
