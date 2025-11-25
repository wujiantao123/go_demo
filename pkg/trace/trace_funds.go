package trace

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

// RPC池
var rpcList = []string{
	"https://api.mainnet-beta.solana.com",
	"https://solana-api.projectserum.com",
	"https://rpc.ankr.com/solana",
	"https://rpc.blockdaemon.com/solana/mainnet",
}

// 提前初始化客户端
var rpcClients []*rpc.Client

func init() {
	for _, endpoint := range rpcList {
		rpcClients = append(rpcClients, rpc.New(endpoint))
	}
}

// 使用 client 轮询 RPC 获取交易
func fetchTransaction(sig string) (*rpc.GetTransactionResult, string, error) {
	version := uint64(0)
	for i, client := range rpcClients {
		tx, err := client.GetTransaction(
			context.Background(),
			solana.MustSignatureFromBase58(sig),
			&rpc.GetTransactionOpts{
				Commitment:                     rpc.CommitmentConfirmed,
				Encoding:                       solana.EncodingBase64,
				MaxSupportedTransactionVersion: &version,
			},
		)
		if err != nil {
			continue
		}
		return tx, rpcList[i], nil
	}
	return nil, "", fmt.Errorf("all RPC failed")
}
func parseSystemInstructions(tx *solana.Transaction) {
	for i, ix := range tx.Message.Instructions {
		if !tx.Message.AccountKeys[ix.ProgramIDIndex].Equals(system.ProgramID) {
			continue
		}

		// Parse system transfer instruction manually
		if len(ix.Data) >= 12 && ix.Data[0] == 2 && len(ix.Accounts) >= 2 {
			// System transfer instruction (instruction type 2)
			from := tx.Message.AccountKeys[ix.Accounts[0]]
			to := tx.Message.AccountKeys[ix.Accounts[1]]

			// Extract lamports from instruction data (bytes 4-12)
			lamports := uint64(ix.Data[4]) |
				uint64(ix.Data[5])<<8 |
				uint64(ix.Data[6])<<16 |
				uint64(ix.Data[7])<<24 |
				uint64(ix.Data[8])<<32 |
				uint64(ix.Data[9])<<40 |
				uint64(ix.Data[10])<<48 |
				uint64(ix.Data[11])<<56

			fmt.Printf("转账 #%d\n", i)
			fmt.Printf("  From  : %s\n", from)
			fmt.Printf("  To    : %s\n", to)
			fmt.Printf("  Amount: %d Lamports (%.9f SOL)\n\n",
				lamports, float64(lamports)/1e9)
		}
	}
}
func GetTransaction() {
	sig := "VwyR57Y3ELc7C5AMTaQ3kfkZjMnFkmEXwvtFkNKF2DZwisP3LHcDJhkqkVCb3q9eLrP9NUXwse2VSGYxk6agu6E"

	tx, used, err := fetchTransaction(sig)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("使用的 RPC:", used)
	transaction, err := tx.Transaction.GetTransaction()
	if err != nil {
		fmt.Println("Error parsing transaction:", err)
		return
	}
	// 解析这笔交易看是否为转账
	fmt.Printf("Transaction: %+v\n", transaction)
	parseSystemInstructions(transaction)
}
