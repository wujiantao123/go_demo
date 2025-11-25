package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FeishuRequest struct {
	MsgType string `json:"msg_type"`
	Content string `json:"content"`
}

func SendMessage(msg string) {
	postMessage := func() error {
		// 获取上海时间
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return err
		}
		shanghaiTime := time.Now().In(loc).Format("2006-01-02 15:04:05")

		// 构建消息体
		contentMap := map[string]string{
			"text": fmt.Sprintf("ca: %s \n%s", msg, shanghaiTime),
		}
		contentJSON, err := json.Marshal(contentMap)
		if err != nil {
			return err
		}

		reqBody := FeishuRequest{
			MsgType: "text",
			Content: string(contentJSON),
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}

		resp, err := http.Post(
			"https://open.feishu.cn/open-apis/bot/v2/hook/48efcce6-6b52-456f-859b-891446fb2995",
			"application/json",
			bytes.NewBuffer(bodyBytes),
		)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("status code: %d", resp.StatusCode)
		}

		return nil
	}

	// 尝试发送消息
	if err := postMessage(); err != nil {
		fmt.Println("发送失败，5秒后重试:", err)
		time.Sleep(5 * time.Second)
		if retryErr := postMessage(); retryErr != nil {
			fmt.Println("重试仍然失败:", retryErr)
		}
	}
}
