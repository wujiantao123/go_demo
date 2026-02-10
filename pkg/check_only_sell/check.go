package checkonlysell

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type CopyTradeConfig struct {
	UserId    string `json:"userId"`
	Username  string `json:"username"`
	PublicKey string `json:"publicKey"`
	Id        string `json:"id"`
	Tag       string `json:"tag"`
	Target    string `json:"target"`

	Reverse         bool  `json:"reverse"`
	ReverseFixedBuy int64 `json:"reverseFixedBuy"`
	Ratio           int   `json:"ratio"`

	TotalUpperLimit            int64 `json:"totalUpperLimit"`
	LowerLimitOfOneTransaction int64 `json:"lowerLimitOfOneTransaction"`
	UpperLimitOfOneTransaction int64 `json:"upperLimitOfOneTransaction"`

	BuyOnce                bool `json:"buyOnce"`
	IgnoreUnburnedLpTokens bool `json:"ignoreUnburnedLpTokens"`
	CopySell               bool `json:"copySell"`

	Slippage     int `json:"slippage"`
	SlippageSell int `json:"slippageSell"`

	PriorityFee     int64 `json:"priorityFee"`
	JitoFee         int64 `json:"jitoFee"`
	PriorityFeeSell int64 `json:"priorityFeeSell"`
	JitoFeeSell     int64 `json:"jitoFeeSell"`

	Enabled bool `json:"enabled"`

	CopyRaydium           bool `json:"copyRaydium"`
	CopyPumpfun           bool `json:"copyPumpfun"`
	CopyPumpfunMayhem     bool `json:"copyPumpfunMayhem"`
	CopyMeteora           bool `json:"copyMeteora"`
	CopyMoonshot          bool `json:"copyMoonshot"`
	CopyPumpamm           bool `json:"copyPumpamm"`
	CopyRaydiumCpmm       bool `json:"copyRaydiumCpmm"`
	CopyRaydiumClmm       bool `json:"copyRaydiumClmm"`
	CopyRaydiumLaunchlab  bool `json:"copyRaydiumLaunchlab"`
	CopyMeteoraDyn        bool `json:"copyMeteoraDyn"`
	CopyMeteoraDbc        bool `json:"copyMeteoraDbc"`
	CopyMeteoraDammv2     bool `json:"copyMeteoraDammv2"`
	CopyBoopfun           bool `json:"copyBoopfun"`
	CopyVertigo           bool `json:"copyVertigo"`
	CopyGavel             bool `json:"copyGavel"`
	CopyPancake           bool `json:"copyPancake"`
	CopyHeaven            bool `json:"copyHeaven"`
	CopyOrca              bool `json:"copyOrca"`
	CopyJupiterAggregator bool `json:"copyJupiterAggregator"`
	CopyOkxAggregator     bool `json:"copyOkxAggregator"`

	MinTokenAge int64 `json:"minTokenAge"`
	MaxTokenAge int64 `json:"maxTokenAge"`
	MinLp       int64 `json:"minLp"`
	MinMc       int64 `json:"minMc"`
	MaxMc       int64 `json:"maxMc"`

	BuyTimes               int  `json:"buyTimes"`
	BuyTimesResetAfterSold bool `json:"buyTimesResetAfterSold"`

	SellByPositionProportion bool `json:"sellByPositionProportion"`
	NotifyNoHolding          bool `json:"notifyNoHolding"`
	RetryTimes               int  `json:"retryTimes"`

	AutoSell       bool   `json:"autoSell"`
	AutoSellTime   int64  `json:"autoSellTime"`
	AutoSellParams string `json:"autoSellParams"`

	IgnoreUnrenouncedLpTokens bool `json:"ignoreUnrenouncedLpTokens"`

	EnableMev     int `json:"enableMev"`
	EnableMevSell int `json:"enableMevSell"`

	ActiveStartTime int64 `json:"activeStartTime"`
	ActiveEndTime   int64 `json:"activeEndTime"`

	PumpfunSlippageTimes int `json:"pumpfunSlippageTimes"`
	SlippagePumpSell     int `json:"slippagePumpSell"`

	FirstSellPercent int `json:"firstSellPercent"`

	TargetSolMaxBuy int64 `json:"targetSolMaxBuy"`
	TargetSolMinBuy int64 `json:"targetSolMinBuy"`

	OnlySell                bool `json:"onlySell"`
	NotCopyPositionAddition bool `json:"notCopyPositionAddition"`

	InviteCode string `json:"inviteCode"`

	EnableTrailingStop   bool   `json:"enableTrailingStop"`
	TrailingStopSettings string `json:"trailingStopSettings"`

	EnableDevSell bool `json:"enableDevSell"`
	DevSellType   int  `json:"devSellType"`
	DevSellBps    int  `json:"devSellBps"`

	EnableTriggerDuration   bool   `json:"enableTriggerDuration"`
	TriggerDurationSettings string `json:"triggerDurationSettings"`

	EnableTurbo bool `json:"enableTurbo"`

	Source            int  `json:"source"`
	SellAllOnTransfer bool `json:"sellAllOnTransfer"`

	IsBuyDip       bool `json:"isBuyDip"`
	TraderHolding  int  `json:"traderHolding"`
	BuyDipDuration int  `json:"buyDipDuration"`
	BuyDipChange   int  `json:"buyDipChange"`

	PvpAnti              bool `json:"pvpAnti"`
	PvpIntervalSec       int  `json:"pvpIntervalSec"`
	PauseCopyDurationSec int  `json:"pauseCopyDurationSec"`

	OnlyCopyDevCreate bool `json:"onlyCopyDevCreate"`
}

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Config CopyTradeConfig `json:"config"`
	} `json:"data"`
}

const token = "l87lfHJBZb5N+5Eoz1hwXzX72QWAIHzGcuZ2HXnfalaVaa4Hj5E06QIi8WVGcC9WXlpKCuR8z0dLEVwJP5JUXH89oB88BdiTs6Qe8mMMvOyqPkhaJSZLI/zKjkRJ58hPp6g5zI/KPDmAFxL6k4KddLdZVGw19rwO6tS4y7K3QNkPOjKUgVp0XjjB/FA4051uUhvwaHceFgi5vsYhG9z1WjLB9k7s0c8Mnrmp6ZO3+Ql+myq4OVhZf6Pnq8wDjT/w+yPxD+P2jEYO81biPzoXValESpcQ4Nsgpjy19tcwEWJKj8nBTR6m7owMx2uDKE88x2hEX1YySITNeygbkE8MFg=="

func CheckOnlySell() {
	url := "https://copy.fastradewiz.com/api/v1/getCopyTrading/538fc78b090d44cbbc7ca8d83de9b7bd"

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("创建请求失败: %w", err)
		return
	}

	// 设置 Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %w", err)
		return
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API 返回错误状态码: %d", resp.StatusCode)
		return
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %w", err)
		return
	}

	// 解析 JSON 响应
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// 如果解析失败，打印原始响应以便调试
		fmt.Printf("JSON 解析失败，原始响应: %s\n", string(body))
		fmt.Printf("解析 JSON 失败: %w", err)
		return
	}

	// 检查 API 返回的 code
	if apiResp.Code != 200 {
		fmt.Printf("API 返回错误: code=%d, message=%s", apiResp.Code, apiResp.Message)
		return
	}

	config := &apiResp.Data.Config
	if config.OnlySell {
		fmt.Println("OnlySell: true，正在更新为 false...")
		config.OnlySell = false
		if err := updateConfig(config); err != nil {
			fmt.Printf("更新 config 失败: %v\n", err)
			return
		}
		fmt.Println("已更新，OnlySell: false")
	} else {
		fmt.Println("OnlySell: false")
	}
}

func updateConfig(config *CopyTradeConfig) error {
	body, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化 config: %w", err)
	}

	url := "https://copy.fastradewiz.com/api/v1/upsertCopyTrading"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回状态码 %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
