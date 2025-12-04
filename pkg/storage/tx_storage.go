package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// FilteredTransaction 过滤后的交易信息
type FilteredTransaction struct {
	Signature string  `json:"signature"`
	Timestamp int64   `json:"timestamp"`
	GasFee    uint64  `json:"gas_fee_lamports"`
	GasFeeSOL float64 `json:"gas_fee_sol"`
	Address   string  `json:"address"`
}

var (
	storageMutex   sync.Mutex
	seenSignatures = make(map[string]bool)
)

// SaveFilteredTransaction 保存符合条件的交易到文件
func SaveFilteredTransaction(tx *FilteredTransaction, filePath string) error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	// 去重检查
	if seenSignatures[tx.Signature] {
		return nil // 已存在，跳过
	}
	seenSignatures[tx.Signature] = true

	// 打开文件（追加模式）
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// JSON Lines 格式：每行一个 JSON 对象
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(tx); err != nil {
		return fmt.Errorf("编码失败: %v", err)
	}

	return nil
}

// GetStoragePath 获取存储文件路径
func GetStoragePath() string {
	// 使用日期作为文件名
	now := time.Now()
	dateStr := now.Format("20060102")
	return fmt.Sprintf("pkg/filtered_tx_%s.jsonl", dateStr)
}
