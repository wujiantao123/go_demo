package close_account

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type tokenAccount struct {
	Lamports   uint64
	Owner      solana.PublicKey
	Data       *accountData
	Executable bool
	Pubkey     solana.PublicKey
	TokenType  string
}

type accountData struct {
	Parsed struct {
		Info struct {
			IsNative    bool   `json:"isNative"`
			Mint        string `json:"mint"`
			Owner       string `json:"owner"`
			State       string `json:"state"`
			TokenAmount struct {
				Amount         string  `json:"amount"`
				Decimals       int     `json:"decimals"`
				UiAmount       float64 `json:"uiAmount"`
				UiAmountString string  `json:"uiAmountString"`
			} `json:"tokenAmount"`
		} `json:"info"`
		Type string `json:"type"`
	} `json:"parsed"`
	Program string `json:"program"`
	Space   int    `json:"space"`
}

func GetAccount() {
	owner := solana.MustPublicKeyFromBase58("2eB41ffTHiyneh4q9DN2XZtE7fvFMQJ21GNNaExNNJTq")

	client := rpc.New("https://mainnet.helius-rpc.com/?api-key=8b7d781c-41a4-464a-9c28-d243fa4b4490")

	// 查询 token accounts (Token2022 + Token)
	resp, err := client.GetTokenAccountsByOwner(
		context.Background(),
		owner,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.Token2022ProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{
			Commitment: rpc.CommitmentFinalized,
			Encoding:   solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		log.Fatal("查询失败: ", err)
	}

	resp2, err := client.GetTokenAccountsByOwner(
		context.Background(),
		owner,
		&rpc.GetTokenAccountsConfig{
			ProgramId: solana.TokenProgramID.ToPointer(),
		},
		&rpc.GetTokenAccountsOpts{
			Commitment: rpc.CommitmentFinalized,
			Encoding:   solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		log.Fatal("查询失败: ", err)
	}

	resp.Value = append(resp.Value, resp2.Value...)

	var closeAccounts []*tokenAccount
	var totalLamports uint64 = 0
	for _, acc := range resp.Value {
		var data accountData
		if err := json.Unmarshal(acc.Account.Data.GetRawJSON(), &data); err != nil {
			log.Println("无法解析账户数据: ", err)
			continue
		}
		if data.Parsed.Info.TokenAmount.UiAmount == 0 {
			closeAccounts = append(closeAccounts, &tokenAccount{
				Lamports:  acc.Account.Lamports,
				Pubkey:    acc.Pubkey,
				TokenType: "token",
			})
			totalLamports += acc.Account.Lamports
		}
	}

	fmt.Println("可关闭 Token Account 数量：", len(closeAccounts))

	totalSOL := float64(totalLamports) / 1_000_000_000
	fmt.Printf("预估可回收 SOL：%.9f SOL\n", totalSOL)
}
