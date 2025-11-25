package trace

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

type TransferInfo struct {
	From     solana.PublicKey
	To       solana.PublicKey
	Lamports uint64
}

// RPC list
var rpcList = []string{
	"https://mainnet.helius-rpc.com/?api-key=8b7d781c-41a4-464a-9c28-d243fa4b4490",
	"https://mainnet.helius-rpc.com/?api-key=c64adbb9-8f0e-48b5-8690-a4d8bb4e5486",
	"https://mainnet.helius-rpc.com/?api-key=fa81dd0b-76fc-434b-83d6-48f151e2d3e5",
	"https://mainnet.helius-rpc.com/?api-key=14312756-eebe-4d84-9617-59a09fc8c894",
	"https://mainnet.helius-rpc.com/?api-key=c570abef-cd38-40b5-a7d8-c599769f7309",
}

var rpcClients []*rpc.Client

func init() {
	for _, endpoint := range rpcList {
		rpcClients = append(rpcClients, rpc.New(endpoint))
	}
}

// 获取余额
func getBalance(addr solana.PublicKey) (uint64, error) {
	for _, c := range rpcClients {
		bal, err := c.GetBalance(context.Background(), addr, rpc.CommitmentConfirmed)
		if err == nil {
			return uint64(bal.Value), nil
		}
	}
	return 0, fmt.Errorf("所有 RPC 获取余额失败")
}

// 获取最近 N 笔交易
func getRecentSignatures(addr solana.PublicKey, limit uint64) ([]rpc.TransactionSignature, error) {
	l := int(limit)

	const (
		maxTries   = 6                      // 总尝试次数（比如有3个RPC，就尝试两轮）
		retryDelay = time.Millisecond * 300 // 每次尝试的间隔
		rpcTimeout = time.Second * 3        // 单个RPC的超时
	)

	rpcCount := len(rpcClients)
	if rpcCount == 0 {
		return nil, fmt.Errorf("没有可用的 RPC 客户端")
	}

	for attempt := 0; attempt < maxTries; attempt++ {
		// 根据 attempt 轮询 RPC
		idx := attempt % rpcCount
		c := rpcClients[idx]

		ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)

		out, err := c.GetSignaturesForAddressWithOpts(
			ctx,
			addr,
			&rpc.GetSignaturesForAddressOpts{
				Limit:      &l,
				Commitment: rpc.CommitmentConfirmed,
			},
		)
		cancel()

		if err == nil {
			// []*T → []T
			res := make([]rpc.TransactionSignature, len(out))
			for i, v := range out {
				if v != nil {
					res[i] = *v
				}
			}
			return res, nil
		}

		fmt.Printf("RPC %s 失败: %v\n", rpcList[idx], err)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("所有 RPC 尝试 %d 次均失败", maxTries)
}

// 获取完整交易
func fetchTransaction(sig string) (*solana.Transaction, error) {
	version := uint64(0)

	for _, c := range rpcClients {
		tx, err := c.GetTransaction(
			context.Background(),
			solana.MustSignatureFromBase58(sig),
			&rpc.GetTransactionOpts{
				Commitment:                     rpc.CommitmentConfirmed,
				Encoding:                       solana.EncodingBase64,
				MaxSupportedTransactionVersion: &version,
			},
		)

		if err == nil && tx != nil && tx.Transaction != nil {
			return tx.Transaction.GetTransaction()
		}
		fmt.Println("RPC 获取交易失败：", err)
	}
	return nil, fmt.Errorf("无法获取交易")
}

// 解析系统转账
func parseSystemInstruction(tx *solana.Transaction) *TransferInfo {
	for _, ix := range tx.Message.Instructions {

		if !tx.Message.AccountKeys[ix.ProgramIDIndex].Equals(system.ProgramID) {
			continue
		}

		if len(ix.Data) >= 12 && ix.Data[0] == 2 && len(ix.Accounts) >= 2 {
			from := tx.Message.AccountKeys[ix.Accounts[0]]
			to := tx.Message.AccountKeys[ix.Accounts[1]]
			lamports := binary.LittleEndian.Uint64(ix.Data[4:12])

			return &TransferInfo{
				From:     from,
				To:       to,
				Lamports: lamports,
			}
		}
	}
	return nil
}

// ------------------------
// 🧠 核心：追踪资金 A→B→C 链路
// ------------------------
func TraceFlow(address solana.PublicKey, depth int, visited map[string]bool) {
	if visited[address.String()] {
		return // 防止循环
	}
	visited[address.String()] = true

	// 读取余额
	bal, _ := getBalance(address)
	fmt.Printf("\n地址 %s 余额：%.9f SOL\n", address, float64(bal)/1e9)
	const MinBalance uint64 = 100000000
	if bal > 0 && bal > MinBalance {
		fmt.Println("余额大于 0或者大于阈值，停止追踪。")
		return
	}

	// 获取最近 10 笔交易
	sigs, err := getRecentSignatures(address, 10)
	if err != nil || len(sigs) == 0 {
		fmt.Printf("无法获取交易或没有交易。%s\n", err)
		return
	}

	for _, sig := range sigs {
		tx, err := fetchTransaction(sig.Signature.String())
		if err != nil {
			continue
		}

		info := parseSystemInstruction(tx)
		if info == nil {
			continue
		}

		// 我们只关心 address 作为 From 的情况
		if info.From.Equals(address) {
			toLink := fmt.Sprintf("https://gmgn.ai/sol/address/%s", info.To)
			fmt.Printf("路径发现：%s → %s (%.9f SOL)\n",
				info.From, toLink, float64(info.Lamports)/1e9)

			SendMessage(fmt.Sprintf(
				"路径发现：%s → %s (%.9f SOL)\n",
				info.From, toLink, float64(info.Lamports)/1e9,
			))

			// 递归追踪 To
			// if depth < 5 { // 最大追踪深度 5
			time.Sleep(time.Millisecond * 300)
			TraceFlow(info.To, depth+1, visited)
			// }
			return
		}
	}

	fmt.Println("找不到从该地址发出的转账。")
}

// 对外调用入口
func StartTrace(addr string) {
	pub := solana.MustPublicKeyFromBase58(addr)
	TraceFlow(pub, 0, map[string]bool{})
}
