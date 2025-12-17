package Stream

import (
	"context"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"log"
	"strings"
	"time"

	"go_demo/pkg/storage"
	"go_demo/pkg/trace"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type YellowstoneGrpcClient struct {
	grpcServer string
	conn       *grpc.ClientConn
}

type TransferInfo struct {
	Kind      string // "transfer" 或 "withdraw-nonce"
	From      string
	To        string
	Lamports  float64
	Authority string // WithdrawNonceAccount 的授权者
}

func NewYellowstoneGrpcClient(grpcServer string) *YellowstoneGrpcClient {
	return &YellowstoneGrpcClient{grpcServer: grpcServer}
}

func (c *YellowstoneGrpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second,
	Timeout:             time.Second,
	PermitWithoutStream: true,
}

// grpc_connect 返回标准 *grpc.ClientConn
func grpc_connect(address string, plaintext bool) *grpc.ClientConn {
	var opts []grpc.DialOption
	if plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, _ := x509.SystemCertPool()
		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	log.Println("Starting grpc client, connecting to", address)
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	return conn
}

// ----------------------- 订阅交易 -----------------------
func (c *YellowstoneGrpcClient) SubscribeTransactions(accounts []string) (*TransactionStream, error) {
	insecureConn := true
	if strings.HasSuffix(c.grpcServer, ":443") {
		insecureConn = false
	}
	c.conn = grpc_connect(c.grpcServer, insecureConn)
	client := pb.NewGeyserClient(c.conn)

	// 建立双向流
	stream, err := client.Subscribe(context.Background())
	if err != nil {
		return nil, err
	}

	failedTransactions := false
	voteTransactions := false
	req := &pb.SubscribeRequest{
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"transactions_sub": {
				Failed:         &failedTransactions,
				Vote:           &voteTransactions,
				AccountInclude: accounts,
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return nil, err
	}

	return &TransactionStream{stream: stream}, nil
}

type TransactionStream struct {
	stream pb.Geyser_SubscribeClient
}

func (s *TransactionStream) Recv() (*pb.SubscribeUpdate, error) {
	return s.stream.Recv()
}

// ----------------------- 订阅账户 -----------------------
func (c *YellowstoneGrpcClient) SubscribeAccounts(accounts []string) (*AccountStream, error) {
	insecureConn := true
	if strings.HasSuffix(c.grpcServer, ":443") {
		insecureConn = false
	}
	c.conn = grpc_connect(c.grpcServer, insecureConn)
	client := pb.NewGeyserClient(c.conn)

	stream, err := client.Subscribe(context.Background())
	if err != nil {
		return nil, err
	}

	req := &pb.SubscribeRequest{
		Accounts: map[string]*pb.SubscribeRequestFilterAccounts{
			"account_sub": {
				Account: accounts,
			},
		},
	}

	if err := stream.Send(req); err != nil {
		return nil, err
	}

	return &AccountStream{stream: stream}, nil
}

type AccountStream struct {
	stream pb.Geyser_SubscribeClient
}

func (s *AccountStream) Recv() (*pb.SubscribeUpdate, error) {
	return s.stream.Recv()
}

// ----------------------- 订阅区块信息 -----------------------
func (c *YellowstoneGrpcClient) SubscribeBlockMeta() (*BlockMetaStream, error) {
	insecureConn := true
	if strings.HasSuffix(c.grpcServer, ":443") {
		insecureConn = false
	}
	c.conn = grpc_connect(c.grpcServer, insecureConn)
	client := pb.NewGeyserClient(c.conn)

	stream, err := client.Subscribe(context.Background())
	if err != nil {
		return nil, err
	}

	req := &pb.SubscribeRequest{
		BlocksMeta: map[string]*pb.SubscribeRequestFilterBlocksMeta{
			"block_meta": {},
		},
	}

	if err := stream.Send(req); err != nil {
		return nil, err
	}

	return &BlockMetaStream{stream: stream}, nil
}

type BlockMetaStream struct {
	stream pb.Geyser_SubscribeClient
}

func (s *BlockMetaStream) Recv() (*pb.SubscribeUpdate, error) {
	return s.stream.Recv()
}

func ParseSystemInstructions(tx *pb.Transaction, signature solana.Signature) []*TransferInfo {
	if tx == nil || tx.Message == nil {
		return []*TransferInfo{}
	}
	if len(tx.Signatures) > 1 {
		// fmt.Println("多签交易，跳过解析", signature.String())
		return []*TransferInfo{}
	}

	var infos []*TransferInfo

	for _, ix := range tx.Message.Instructions {
		if len(ix.Data) < 1 || len(ix.Accounts) < 2 {
			continue
		}

		program := tx.Message.AccountKeys[ix.ProgramIdIndex]
		pk := solana.PublicKeyFromBytes(program)
		if !pk.Equals(system.ProgramID) {
			continue
		}
		instr := ix.Data[0]
		switch instr {
		// case 2: // Transfer
		// 	if len(ix.Data) < 12 || len(ix.Accounts) < 2 {
		// 		continue
		// 	}

		// 	lamports := binary.LittleEndian.Uint64(ix.Data[4:12])
		// 	from := tx.Message.AccountKeys[ix.Accounts[0]]
		// 	to := tx.Message.AccountKeys[ix.Accounts[1]]

		// 	infos = append(infos, &TransferInfo{
		// 		Kind:     "transfer",
		// 		From:     solana.PublicKeyFromBytes(from).String(),
		// 		To:       solana.PublicKeyFromBytes(to).String(),
		// 		Lamports: float64(lamports) / 1e9,
		// 	})

		case 5: // WithdrawNonceAccount
			if len(ix.Data) < 9 || len(ix.Accounts) < 5 {
				continue
			}

			lamports := binary.LittleEndian.Uint64(ix.Data[1:9])
			nonceAcc := tx.Message.AccountKeys[ix.Accounts[0]]
			dest := tx.Message.AccountKeys[ix.Accounts[1]]
			auth := tx.Message.AccountKeys[ix.Accounts[4]]

			infos = append(infos, &TransferInfo{
				Kind:      "withdraw-nonce",
				From:      solana.PublicKeyFromBytes(nonceAcc).String(),
				To:        solana.PublicKeyFromBytes(dest).String(),
				Lamports:  float64(lamports) / 1e9,
				Authority: solana.PublicKeyFromBytes(auth).String(),
			})
		default:
			continue
		}
	}

	return infos
}
func Subscribe() {
	client := NewYellowstoneGrpcClient("84.32.103.140:10030")
	defer client.Close()

	for {
		// 建立流
		txStream, err := client.SubscribeTransactions([]string{})
		if err != nil {
			log.Printf("SubscribeTransactions error: %v, retry in 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("Subscribed to transaction stream")

		for {
			update, err := txStream.Recv()
			if err != nil {
				log.Printf("Error receiving transaction: %v, reconnecting in 3s...", err)
				time.Sleep(3 * time.Second)
				break // 跳出内层循环，重新订阅
			}

			transaction, ok := update.UpdateOneof.(*pb.SubscribeUpdate_Transaction)
			if !ok {
				continue
			}

			txn := transaction.Transaction.Transaction
			// signer := solana.PublicKeyFromBytes(txn.Transaction.Message.AccountKeys[0])
			signature := solana.SignatureFromBytes(txn.Signature)
			// fmt.Printf("Transaction Signature: %s, Signer: %s\n", signature.String(), signer.String())

			infos := ParseSystemInstructions(txn.Transaction, signature)
			if len(infos) > 0 {
				for _, info := range infos {
					fmt.Printf(
						"  Type: %s, From: %s, To: %s, Lamports: %.9f, Authority: %s Signature: %s\n",
						info.Kind,
						info.From,
						info.To,
						info.Lamports,
						info.Authority,
						signature.String(),
					)
				}

			}
		}
	}
}

// SubscribeFilteredTransactions 订阅并过滤交易
func SubscribeFilteredTransactions() {
	// 目标地址
	targetAddresses := []solana.PublicKey{
		// solana.MustPublicKeyFromBase58("b1oomGGqPKGD6errbyfbVMBuzSC8WtAAYo8MwNafWW1"),
		// solana.MustPublicKeyFromBase58("3kxSQybWEeQZsMuNWMRJH4TxrhwoDwfv41TNMLRzFP5A"),
	}
	minGasFeeSOL := 0.001

	client := NewYellowstoneGrpcClient("84.32.103.140:10030")
	defer client.Close()

	storagePath := storage.GetStoragePath()
	log.Printf("开始监听交易，符合条件的将保存到: %s", storagePath)

	for {
		// 建立流 - 如果 gRPC 端已经过滤了这两个地址，可以传入
		accountsFilter := []string{
			"6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P",
			// "b1oomGGqPKGD6errbyfbVMBuzSC8WtAAYo8MwNafWW1",
			// "3kxSQybWEeQZsMuNWMRJH4TxrhwoDwfv41TNMLRzFP5A",
		}
		txStream, err := client.SubscribeTransactions(accountsFilter)
		if err != nil {
			log.Printf("SubscribeTransactions error: %v, retry in 3s...", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("Subscribed to transaction stream")

		for {
			update, err := txStream.Recv()
			if err != nil {
				log.Printf("Error receiving transaction: %v, reconnecting in 3s...", err)
				time.Sleep(3 * time.Second)
				break
			}

			transaction, ok := update.UpdateOneof.(*pb.SubscribeUpdate_Transaction)
			if !ok {
				continue
			}

			isValid, txInfo, err := trace.CheckTransactionFromGrpc(transaction.Transaction, targetAddresses, minGasFeeSOL)
			if err != nil {
				log.Printf("检查交易失败: %v", err)
				continue
			}

			if isValid && txInfo != nil {
				log.Printf("✅ 找到符合条件的交易: %s, Gas Fee: %.9f SOL", txInfo.Signature, txInfo.GasFeeSOL)

				// 保存到文件
				filteredTx := &storage.FilteredTransaction{
					Signature: txInfo.Signature,
					Timestamp: time.Now().Unix(),
					GasFee:    txInfo.GasFee,
					GasFeeSOL: txInfo.GasFeeSOL,
					Address:   txInfo.Address,
				}

				if err := storage.SaveFilteredTransaction(filteredTx, storagePath); err != nil {
					log.Printf("保存交易失败 %s: %v", txInfo.Signature, err)
				} else {
					log.Printf("已保存交易: %s", txInfo.Signature)
				}
			}
		}
	}
}
