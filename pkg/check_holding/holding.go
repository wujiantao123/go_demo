package checkholding

import (
	"fmt"
	"math/rand"
	"time"

	"go_demo/pkg/trace"

	"github.com/gagliardetto/solana-go"
)

const (
	addressA      = "CRyD3wPLP41rcFhN9iP1D96SDPYwJiPMWPa25x4TR2tp"
	addressB      = "78MW1gjBE1nGTELzL12tJY614uNqbFLHb4wYxr537NRK"
	pollInterval  = 5 * time.Second // 每 5 秒轮询一次
	dustThreshold = 0.001           // 小于 0.001 不处理
)

// 记录 B 已知持仓
var bKnownMints = make(map[string]float64)

// 查询 B 或 A 的所有 token accounts
func getTokenAccounts(address string) (map[string]float64, error) {
	// 使用 trace 包导出的 RPC 客户端
	rpcClients := trace.GetRPCClients()

	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return nil, fmt.Errorf("无效的地址: %v", err)
	}

	// 尝试从所有 RPC 客户端获取 token accounts
	// TODO: 实现 getTokenAccountsByOwner 来查询 token 持仓
	// 目前先返回空 map 作为占位符
	// 示例：如何使用 RPC 客户端
	// for _, c := range rpcClients {
	//     // 使用 c 进行 RPC 调用
	//     // 例如: c.GetTokenAccountsByOwner(...)
	// }
	_ = rpcClients
	_ = pubKey
	return make(map[string]float64), nil
	// url := getRpcUrl()
	// payload := map[string]interface{}{
	// 	"jsonrpc": "2.0",
	// 	"id":      1,
	// 	"method":  "getTokenAccountsByOwner",
	// 	"params": []interface{}{
	// 		address,
	// 		map[string]string{"programId": "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},
	// 		map[string]string{"encoding": "jsonParsed"},
	// 	},
	// }

	// data, _ := json.Marshal(payload)
	// resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	// if err != nil {
	// 	return nil, err
	// }
	// defer resp.Body.Close()
	// body, _ := ioutil.ReadAll(resp.Body)

	// // 解析 RPC 返回
	// var rpcResp struct {
	// 	Result []struct {
	// 		Account struct {
	// 			Data struct {
	// 				Parsed struct {
	// 					Info struct {
	// 						Mint     string `json:"mint"`
	// 						TokenAmt struct {
	// 							Amount string `json:"amount"`
	// 						} `json:"tokenAmount"`
	// 					} `json:"info"`
	// 				} `json:"parsed"`
	// 			} `json:"data"`
	// 		} `json:"account"`
	// 	} `json:"result"`
	// }

	// if err := json.Unmarshal(body, &rpcResp); err != nil {
	// 	return nil, err
	// }

	// holdings := make(map[string]float64)
	// for _, t := range rpcResp.Result {
	// 	mint := t.Account.Data.Parsed.Info.Mint
	// 	var amt float64
	// 	fmt.Sscanf(t.Account.Data.Parsed.Info.TokenAmt.Amount, "%f", &amt)
	// 	holdings[mint] = amt
	// }
	// return holdings, nil
}

// 卖出 B 的 token（接入你实际交易逻辑）
func forceSellB(mint string, amount float64) {
	fmt.Printf("⚠️ B 有仓但 A 没仓，卖出 B token: %s, 数量: %.6f\n", mint, amount)
	// TODO: 这里接入 Serum / Raydium / Orca 卖出逻辑
}

func StartCheckHolding() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("🚀 被动守护程序启动...")

	for {
		// 查询 A 的持仓
		aHoldings, err := getTokenAccounts(addressA)
		if err != nil {
			fmt.Println("查询 A 持仓失败:", err)
			time.Sleep(pollInterval)
			continue
		}

		// 查询 B 的持仓
		bHoldings, err := getTokenAccounts(addressB)
		if err != nil {
			fmt.Println("查询 B 持仓失败:", err)
			time.Sleep(pollInterval)
			continue
		}

		// 检测 B 新增的 mint
		for mint, bAmt := range bHoldings {
			if bAmt < dustThreshold {
				continue
			}
			oldAmt, known := bKnownMints[mint]
			if !known || oldAmt != bAmt {
				// 更新已知持仓
				bKnownMints[mint] = bAmt

				// 检查 A 是否有仓
				aAmt, ok := aHoldings[mint]
				if !ok || aAmt < dustThreshold {
					// B 有仓但 A 没仓 → 卖出 B
					forceSellB(mint, bAmt)
				}
			}
		}

		time.Sleep(pollInterval)
	}
}
