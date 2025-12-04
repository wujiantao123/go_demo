package trace

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

type TransferInfo struct {
	From      solana.PublicKey
	To        solana.PublicKey
	Lamports  uint64
	Kind      string // transfer / withdraw-nonce
	Authority solana.PublicKey
}

// TraceConfig 追踪配置，包含地址和备注
type TraceConfig struct {
	Address string // 地址
	Remark  string // 备注
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

// 获取完整交易结果（包含 fee 等元数据）
func fetchTransactionResult(sig string) (*rpc.GetTransactionResult, error) {
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

		if err == nil && tx != nil {
			return tx, nil
		}
		fmt.Println("RPC 获取交易失败：", err)
	}
	return nil, fmt.Errorf("无法获取交易")
}

// 解析系统转账
func parseSystemInstruction(tx *solana.Transaction) *TransferInfo {
	if len(tx.Signatures) > 1 {
		fmt.Println("多签交易，跳过解析")
		return nil
	}
	for _, ix := range tx.Message.Instructions {

		// 不是 SystemProgram 跳过
		if !tx.Message.AccountKeys[ix.ProgramIDIndex].Equals(system.ProgramID) {
			continue
		}

		// Data 至少要有 1 byte 指令 index
		if len(ix.Data) < 1 {
			continue
		}

		instr := ix.Data[0]
		fmt.Println("instr:", instr)
		switch instr {

		// -------------------------------------------------------
		// SystemInstruction::Transfer = 2
		// -------------------------------------------------------
		case 2:
			if len(ix.Data) >= 12 && len(ix.Accounts) >= 2 {

				from := tx.Message.AccountKeys[ix.Accounts[0]]
				to := tx.Message.AccountKeys[ix.Accounts[1]]
				lamports := binary.LittleEndian.Uint64(ix.Data[4:12])

				return &TransferInfo{
					From:     from,
					To:       to,
					Lamports: lamports,
					Kind:     "transfer",
				}
			}

		// -------------------------------------------------------
		// SystemInstruction::WithdrawNonceAccount = 7
		// -------------------------------------------------------
		case 5:
			// WithdrawNonceAccount 结构最少要求 5 个 account
			if len(ix.Data) >= 9 && len(ix.Accounts) >= 5 {

				nonceAccount := tx.Message.AccountKeys[ix.Accounts[0]]
				recipient := tx.Message.AccountKeys[ix.Accounts[1]]
				authority := tx.Message.AccountKeys[ix.Accounts[4]]

				lamports := binary.LittleEndian.Uint64(ix.Data[1:9])

				return &TransferInfo{
					From:      nonceAccount,
					To:        recipient,
					Lamports:  lamports,
					Kind:      "withdraw-nonce",
					Authority: authority,
				}
			}
		}
	}

	return nil
}

// ------------------------
// 🧠 核心：追踪资金 A→B→C 链路
// ------------------------
func TraceFlow(address solana.PublicKey, visited map[string]bool, remark string) {
	const MinBalance uint64 = 100_000_000

	for {
		// 读取余额
		bal, _ := getBalance(address)
		fmt.Printf("\n备注 %s 地址 %s 余额：%.9f SOL\n", remark, address, float64(bal)/1e9)

		if bal > 0 && bal > MinBalance {
			fmt.Println("余额大于 0或者大于阈值，继续轮询...")
			// 等待 20 秒再继续检查
			time.Sleep(20 * time.Second)
			continue
		}

		// 获取最近 10 笔交易
		sigs, err := getRecentSignatures(address, 10)
		if err != nil || len(sigs) == 0 {
			fmt.Printf("无法获取交易或没有交易。%s\n", err)
			return
		}

		found := false
		for _, sig := range sigs {
			tx, err := fetchTransaction(sig.Signature.String())
			if err != nil {
				continue
			}
			info := parseSystemInstruction(tx)
			fmt.Println("转账信息：", info)
			fmt.Println("检查地址：", address.String())
			if info == nil {
				continue
			}
			fmt.Printf("路径发现：%s → %s (%.9f SOL)\n", info.From, info.To, float64(info.Lamports)/1e9)

			if info.From.Equals(address) {
				toLink := fmt.Sprintf("https://gmgn.ai/sol/address/%s", info.To)
				fmt.Printf("路径发现：%s → %s (%.9f SOL)\n",
					info.From, toLink, float64(info.Lamports)/1e9)

				// 发送消息时带上备注
				message := fmt.Sprintf(
					"[%s] 路径发现：%s → %s (%.9f SOL)\n",
					remark, info.From, toLink, float64(info.Lamports)/1e9,
				)
				SendMessage(message)

				// 避免递归无限循环
				if !visited[info.To.String()] {
					visited[info.To.String()] = true
					time.Sleep(300 * time.Millisecond)
					TraceFlow(info.To, visited, remark)
				}

				found = true
				break // 找到一次交易就处理完当前轮
			}
		}

		if !found {
			fmt.Println("找不到从该地址发出的转账。")
			return
		}
	}
}

// 对外调用入口
func StartTrace(configs []TraceConfig) {
	for _, config := range configs {
		pub := solana.MustPublicKeyFromBase58(config.Address)
		go TraceFlow(pub, map[string]bool{}, config.Remark)
	}
}

// 检查交易是否包含指定地址
func containsAddress(tx *solana.Transaction, targetAddr solana.PublicKey) bool {
	if tx == nil {
		return false
	}

	// 检查所有账户密钥
	for _, key := range tx.Message.AccountKeys {
		if key.Equals(targetAddr) {
			return true
		}
	}

	// 签名者通常是第一个账户
	if len(tx.Message.AccountKeys) > 0 {
		signer := tx.Message.AccountKeys[0]
		if signer.Equals(targetAddr) {
			return true
		}
	}

	return false
}

// 检查是否是 meme 交易（涉及指定的 token mint）
func isMemeTransaction(tx *solana.Transaction, memeMint solana.PublicKey) bool {
	if tx == nil {
		return false
	}

	// 检查所有账户密钥中是否包含 meme mint 地址
	for _, key := range tx.Message.AccountKeys {
		if key.Equals(memeMint) {
			return true
		}
	}

	// 检查指令中的账户
	for _, ix := range tx.Message.Instructions {
		for _, accIdx := range ix.Accounts {
			if int(accIdx) < len(tx.Message.AccountKeys) {
				if tx.Message.AccountKeys[accIdx].Equals(memeMint) {
					return true
				}
			}
		}
	}

	return false
}

// 获取交易的 gas fee
func getTransactionFee(txResult *rpc.GetTransactionResult) uint64 {
	if txResult == nil || txResult.Meta == nil {
		return 0
	}
	return txResult.Meta.Fee
}

// MemeTransactionInfo meme 交易信息
type MemeTransactionInfo struct {
	Signature     string
	ContainsB1oom bool
	IsMemeTx      bool
	GasFee        uint64
	GasFeeSOL     float64
	Transaction   *solana.Transaction
}

// AnalyzeMemeTransaction 分析交易是否包含指定地址且是 meme 交易
func AnalyzeMemeTransaction(sig string) (*MemeTransactionInfo, error) {
	// 目标地址
	b1oomAddr := solana.MustPublicKeyFromBase58("b1oomGGqPKGD6errbyfbVMBuzSC8WtAAYo8MwNafWW1")
	// Meme token mint 地址
	memeMint := solana.MustPublicKeyFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM")

	// 获取交易结果（包含 fee）
	txResult, err := fetchTransactionResult(sig)
	if err != nil {
		return nil, fmt.Errorf("获取交易失败: %v", err)
	}

	if txResult.Transaction == nil {
		return nil, fmt.Errorf("交易数据为空")
	}

	tx, err := txResult.Transaction.GetTransaction()
	if err != nil {
		return nil, fmt.Errorf("解析交易失败: %v", err)
	}

	// 检查是否包含 b1oom 地址
	containsB1oom := containsAddress(tx, b1oomAddr)

	// 检查是否是 meme 交易
	isMemeTx := isMemeTransaction(tx, memeMint)

	// 获取 gas fee
	gasFee := getTransactionFee(txResult)

	info := &MemeTransactionInfo{
		Signature:     sig,
		ContainsB1oom: containsB1oom,
		IsMemeTx:      isMemeTx,
		GasFee:        gasFee,
		GasFeeSOL:     float64(gasFee) / 1e9,
		Transaction:   tx,
	}

	return info, nil
}

func GetTransaction(sig string) (*solana.Transaction, error) {
	tx, err := fetchTransaction(sig)
	// fmt.Println("交易：", tx)
	info := parseSystemInstruction(tx)
	fmt.Println("转账信息：", info)
	return tx, err
}

// containsAllAddresses 检查交易是否同时包含所有指定的地址
func containsAllAddresses(tx *solana.Transaction, targetAddresses []solana.PublicKey) bool {
	if tx == nil || len(targetAddresses) == 0 {
		return false
	}

	// 将交易中的所有地址转为 map 以便快速查找
	addressMap := make(map[string]bool)
	for _, key := range tx.Message.AccountKeys {
		addressMap[key.String()] = true
	}

	// 检查指令中的账户
	for _, ix := range tx.Message.Instructions {
		for _, accIdx := range ix.Accounts {
			if int(accIdx) < len(tx.Message.AccountKeys) {
				addressMap[tx.Message.AccountKeys[accIdx].String()] = true
			}
		}
	}

	// 检查是否包含所有目标地址
	for _, targetAddr := range targetAddresses {
		if !addressMap[targetAddr.String()] {
			return false
		}
	}

	return true
}

// FilteredTransactionInfo 过滤后的交易信息（用于内部传递）
type FilteredTransactionInfo struct {
	Signature string
	GasFee    uint64
	GasFeeSOL float64
	Address   string
}

// getGasFeeFromSignature 通过签名获取交易的 gas fee
func getGasFeeFromSignature(sig string) (uint64, error) {
	txResult, err := fetchTransactionResult(sig)
	if err != nil {
		return 0, err
	}
	return getTransactionFee(txResult), nil
}

// CheckTransactionFromGrpc 直接从 gRPC 交易信息中检查交易是否符合条件
func CheckTransactionFromGrpc(txnInfo *pb.SubscribeUpdateTransaction, targetAddresses []solana.PublicKey, minGasFeeSOL float64) (bool, *FilteredTransactionInfo, error) {
	if txnInfo == nil || txnInfo.Transaction == nil {
		return false, nil, nil
	}

	pbTx := txnInfo.Transaction.Transaction
	if pbTx == nil || pbTx.Message == nil {
		return false, nil, nil
	}

	// 获取签名
	if len(pbTx.Signatures) == 0 {
		return false, nil, nil
	}
	signature := solana.SignatureFromBytes(pbTx.Signatures[0])
	sigStr := signature.String()

	var accountKeys []solana.PublicKey
	for _, keyBytes := range pbTx.Message.AccountKeys {
		accountKeys = append(accountKeys, solana.PublicKeyFromBytes(keyBytes))
	}

	// 构建地址映射以检查是否包含所有目标地址
	addressMap := make(map[string]bool)
	for _, key := range accountKeys {
		addressMap[key.String()] = true
	}
	// 获取 gas fee - 从 gRPC 消息的 Meta 中获取，如果没有则通过 RPC 查询
	var gasFee uint64
	if txnInfo.Transaction != nil && txnInfo.Transaction.Meta != nil {
		gasFee = txnInfo.Transaction.Meta.Fee
	}
	gasFeeSOL := float64(gasFee) / 1e9
	fmt.Printf("gasFee %.9f 交易 %s \n", gasFeeSOL, sigStr)

	// 检查指令中的账户
	for _, ix := range pbTx.Message.Instructions {
		for _, accIdx := range ix.Accounts {
			if int(accIdx) < len(accountKeys) {
				addressMap[accountKeys[accIdx].String()] = true
			}
		}
	}

	// 检查是否包含所有目标地址
	for _, targetAddr := range targetAddresses {
		if !addressMap[targetAddr.String()] {
			return false, nil, nil
		}
	}

	if gasFeeSOL <= minGasFeeSOL {
		return false, nil, nil
	}

	info := &FilteredTransactionInfo{
		Signature: sigStr,
		GasFee:    gasFee,
		GasFeeSOL: gasFeeSOL,
		Address:   accountKeys[0].String(),
	}

	return true, info, nil
}
