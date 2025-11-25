package trace

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

type TransferInfo struct {
	From     solana.PublicKey
	To       solana.PublicKey
	Lamports uint64
}

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

func parseSystemInstructions(tx *solana.Transaction) *TransferInfo {
	for _, ix := range tx.Message.Instructions {

		// 不是 system program → 跳过
		if !tx.Message.AccountKeys[ix.ProgramIDIndex].Equals(system.ProgramID) {
			continue
		}

		// System Transfer 指令
		// ix.Data[0] == 2 说明是 transfer
		if len(ix.Data) >= 12 && ix.Data[0] == 2 && len(ix.Accounts) >= 2 {

			from := tx.Message.AccountKeys[ix.Accounts[0]]
			to := tx.Message.AccountKeys[ix.Accounts[1]]

			// lamports = u64，从 byte 4 开始连续 8 字节
			lamports := binary.LittleEndian.Uint64(ix.Data[4:12])

			return &TransferInfo{
				From:     from,
				To:       to,
				Lamports: lamports,
			}
		}
	}

	// 如果没有转账指令 → 返回 nil
	return nil
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
	info := parseSystemInstructions(transaction)
	if info != nil {
		fmt.Println("发生转账：")
		fmt.Println("From:", info.From)
		fmt.Println("To:", info.To)
		fmt.Printf("Amount: %d (%.9f SOL)\n",
			info.Lamports, float64(info.Lamports)/1e9)
	}
}
