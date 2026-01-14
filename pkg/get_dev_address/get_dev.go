package get_dev_address

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

// APIResponse 表示 API 返回的数据结构
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Data []struct {
			TokenAddress string `json:"token_address"`
		} `json:"data"`
	} `json:"data"`
}

// TokenDetailResponse 表示 token 详情 API 返回的数据结构
type TokenDetailResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	TraceID string `json:"traceId"`
	Data    struct {
		Creator string `json:"creator"`
		// 可以根据需要添加其他字段
	} `json:"data"`
}

// UpsertCopyTradingResponse 表示添加复制交易 API 返回的数据结构
type UpsertCopyTradingResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type AddAKAddressResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data int    `json:"data"`
}

// GetCopyTradingListResponse 表示获取复制交易列表 API 返回的数据结构
type GetCopyTradingListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Config                   interface{} `json:"config"`
		CopyTradingListPnlParams interface{} `json:"copyTradingListPnlParams"`
		Count                    int         `json:"count"`
		List                     []struct {
			Target    string `json:"target"`
			Publickey string `json:"publickey"`
			// 可以根据需要添加其他字段
		} `json:"list"`
	} `json:"data"`
}

// GetDevAddress 获取所有 dev 的地址
func GetAddressTokenList() ([]string, error) {
	// url := "https://copy.fastradewiz.com/api/v1/address/token_pnl_list?limit=500&address=H1XEpF9HHVoqDzrjrwbhvgr75rmDhGJDJHmRPr8ZeFwK&sort_by=last_trade_time&sort_order=desc&holding=false&next="
	url := "https://copy.fastradewiz.com/api/v1/address/token_pnl_list?limit=500&address=DDDDHjHoSWQ8TSZdWJnaxBaPSPUqVjKNbhJ4MhLXUDDD&sort_by=last_trade_time&sort_order=desc&holding=false&next="
	token := "l87lfHJBZb5N+5Eoz1hwXzX72QWAIHzGcuZ2HXnfalaVaa4Hj5E06QIi8WVGcC9WXlpKCuR8z0dLEVwJP5JUXH89oB88BdiTs6Qe8mMMvOyqPkhaJSZLI/zKjkRJ58hPp6g5zI/KPDmAFxL6k4KddLdZVGw19rwO6tS4y7K3QNkPOjKUgVp0XjjB/FA4051uUhvwaHceFgi5vsYhG9z1WjLB9k7s0c8Mnrmp6ZO3+Ql+myq4OVhZf6Pnq8wDjT/w+yPxD+P2jEYO81biPzoXValESpcQ4Nsgpjy19tcwEWJKj8nBTR6m7owMx2uDKE88x2hEX1YySITNeygbkE8MFg=="

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析 JSON 响应
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// 如果解析失败，打印原始响应以便调试
		fmt.Printf("JSON 解析失败，原始响应: %s\n", string(body))
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 检查 API 返回的 code
	if apiResp.Code != 200 {
		return nil, fmt.Errorf("API 返回错误: code=%d, message=%s", apiResp.Code, apiResp.Message)
	}

	// 提取 token_address 列表
	var addresses []string
	for _, item := range apiResp.Data.Data {
		if item.TokenAddress != "" {
			addresses = append(addresses, item.TokenAddress)
		}
	}
	fmt.Println(addresses)
	return addresses, nil
}

// GetTokenDetail 获取 token 详情并返回 creator 地址
func GetTokenDetail(tokenAddress string) (string, error) {
	url := "https://tradewiz.ai/api/v1/detail"
	method := "POST"

	// 使用传入的 tokenAddress 参数构建 payload
	payloadStr := fmt.Sprintf(`{"token_address":"%s"}`, tokenAddress)
	payload := strings.NewReader(payloadStr)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Cookie", "Cookie=_ga=GA1.1.1043995264.1754480254; login_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJsb2dpbiIsImV4cCI6MTc5NDIxNjg4NCwianRpIjoiMzgifQ.88lmSxdtxzfYoFIuOzti2-drZwV4Kzdn8AL7t9-dd_w; _ga_3XSMJ94W9X=GS2.1.s1768296796$o19$g1$t1768297904$j60$l0$h0")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer res.Body.Close()

	// 检查 HTTP 状态码
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误状态码: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析 JSON 响应
	var detailResp TokenDetailResponse
	if err := json.Unmarshal(body, &detailResp); err != nil {
		// 如果解析失败，打印原始响应以便调试
		fmt.Printf("JSON 解析失败，原始响应: %s\n", string(body))
		return "", fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 检查 API 返回的 code
	if detailResp.Code != 200 {
		return "", fmt.Errorf("API 返回错误: code=%d, msg=%s", detailResp.Code, detailResp.Msg)
	}

	// 返回 creator 地址
	return detailResp.Data.Creator, nil
}

func AddTradewizCopyAddress(address string) error {
	url := "https://copy.fastradewiz.com/api/v1/upsertCopyTrading"
	method := "POST"
	if len(address) < 6 {
		return fmt.Errorf("地址长度小于6")
	}
	tag := "dev-" + address[:6]
	payloadStr := fmt.Sprintf(`{
	 "tag": "%s",
	 "target": "%s",
	 "autoSell": false,
	 "autoSellParams": "{\"settings\":{}}",
	 "autoSellTime": 0,
	 "buyTimes": 1,
	 "buyTimesResetAfterSold": false,
	 "copySell": true,
	 "enableMev": 0,
	 "enableMevSell": 0,
	 "enableTrailingStop": false,
	 "enableTurbo": true,
	 "enabled": false,
	 "firstSellPercent": 0,
	 "ignoreUnburnedLpTokens": false,
	 "ignoreUnrenouncedLpTokens": false,
	 "jitoFee": 2000000,
	 "jitoFeeSell": 0,
	 "lowerLimitOfOneTransaction": -1,
	 "upperLimitOfOneTransaction": 395000000,
	 "totalUpperLimit": -1,
	 "maxMc": 5000,
	 "minMc": -1,
	 "maxTokenAge": -1,
	 "minTokenAge": -1,
	 "minLp": -1,
	 "notCopyPositionAddition": false,
	 "notifyNoHolding": false,
	 "onlySell": false,
	 "priorityFee": 3000000,
	 "priorityFeeSell": 100000,
	 "pumpfunSlippageTimes": 40,
	 "ratio": 100,
	 "retryTimes": 0,
	 "sellByPositionProportion": true,
	 "slippage": 12,
	 "slippageSell": 22,
	 "slippagePumpSell": 30,
	 "targetSolMaxBuy": -1,
	 "targetSolMinBuy": -1,
	 "copyPumpfun": true,
	 "copyPumpfunMayhem": true,
	 "copyRaydiumLaunchlab": true,
	 "copyRaydium": true,
	 "copyRaydiumCpmm": true,
	 "copyRaydiumClmm": true,
	 "copyMeteora": true,
	 "copyMeteoraDbc": true,
	 "copyMeteoraDyn": true,
	 "copyMeteoraDammv2": true,
	 "copyPumpamm": true,
	 "copyJupiterAggregator": true,
	 "copyMoonshot": true,
	 "copyBoopfun": true,
	 "copyGavel": true,
	 "copyVertigo": true,
	 "copyPancake": true,
	 "copyHeaven": true,
	 "copyOkxAggregator": true,
	 "copyOrca": true,
	 "activeStartTime": -1,
	 "activeEndTime": -1,
	 "enableDevSell": false,
	 "devSellType": 0,
	 "devSellBps": 10000,
	 "enableTriggerDuration": true,
	 "publicKey": "2UcSusoUPz1QcSoCovRGpy5G6MGiqPkRRzbiVoGnCevE",
	 "trailingStopSettings": "[]",
	 "triggerDurationSettings": "{\"3\":10000}",
	 "isBuyDip": false,
	 "buyDipChange": 0,
	 "traderHolding": 1,
	 "buyDipDuration": 60,
	 "sellAllOnTransfer": true
 }`, tag, address)
	payload := strings.NewReader(payloadStr)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Authorization", "Bearer l87lfHJBZb5N+5Eoz1hwXzX72QWAIHzGcuZ2HXnfalaVaa4Hj5E06QIi8WVGcC9WXlpKCuR8z0dLEVwJP5JUXH89oB88BdiTs6Qe8mMMvOyqPkhaJSZLI/zKjkRJ58hPp6g5zI/KPDmAFxL6k4KddLdZVGw19rwO6tS4y7K3QNkPOjKUgVp0XjjB/FA4051uUhvwaHceFgi5vsYhG9z1WjLB9k7s0c8Mnrmp6ZO3+Ql+myq4OVhZf6Pnq8wDjT/w+yPxD+P2jEYO81biPzoXValESpcQ4Nsgpjy19tcwEWJKj8nBTR6m7owMx2uDKE88x2hEX1YySITNeygbkE8MFg==")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer res.Body.Close()

	// 检查 HTTP 状态码
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API 返回错误状态码: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析 JSON 响应
	var resp UpsertCopyTradingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		// 如果解析失败，打印原始响应以便调试
		fmt.Printf("JSON 解析失败，原始响应: %s\n", string(body))
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 检查 API 返回的 code
	if resp.Code != 200 {
		return fmt.Errorf("API 返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}

	fmt.Printf("成功添加复制交易地址: %s, 响应: %s\n", address, resp.Data)
	return nil
}

// GetTradewizCopyAddress 获取所有复制交易地址的 target 列表
func GetTradewizCopyAddress() ([]string, error) {
	url := "https://copy.fastradewiz.com/api/v1/getCopyTradingList"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Authorization", "Bearer l87lfHJBZb5N+5Eoz1hwXzX72QWAIHzGcuZ2HXnfalaVaa4Hj5E06QIi8WVGcC9WXlpKCuR8z0dLEVwJP5JUXH89oB88BdiTs6Qe8mMMvOyqPkhaJSZLI/zKjkRJ58hPp6g5zI/KPDmAFxL6k4KddLdZVGw19rwO6tS4y7K3QNkPOjKUgVp0XjjB/FA4051uUhvwaHceFgi5vsYhG9z1WjLB9k7s0c8Mnrmp6ZO3+Ql+myq4OVhZf6Pnq8wDjT/w+yPxD+P2jEYO81biPzoXValESpcQ4Nsgpjy19tcwEWJKj8nBTR6m7owMx2uDKE88x2hEX1YySITNeygbkE8MFg==")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer res.Body.Close()

	// 检查 HTTP 状态码
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回错误状态码: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析 JSON 响应
	var resp GetCopyTradingListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		// 如果解析失败，打印原始响应以便调试
		fmt.Printf("JSON 解析失败，原始响应: %s\n", string(body))
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 检查 API 返回的 code
	if resp.Code != 200 {
		return nil, fmt.Errorf("API 返回错误: code=%d, message=%s", resp.Code, resp.Message)
	}

	// 提取 target 地址列表
	var targets []string
	for _, item := range resp.Data.List {
		if item.Target != "" && item.Publickey == "2UcSusoUPz1QcSoCovRGpy5G6MGiqPkRRzbiVoGnCevE" {
			targets = append(targets, item.Target)
		}
	}

	return targets, nil
}

func AddAKAddress(address string) error {
	url := "https://api.akbot.pro/api/akBot/add"
	method := "POST"
	payloadStr := fmt.Sprintf(`{"akCommonConfig":{"enableSwitch":1,"copyAddressNote":null,"akCopyAddresses":[{"key":"38ed59c5-f6ee-4286-a765-949859eb5806","copyAddress":"%s"}],"akConfigLabels":[],"platform":"pump_fun,pump_amm","groupSlippageMultiple":1,"positionFollowSwitch":1,"firstPurchaseSwitch":1,"copyMinAmount":0,"copyMaxAmount":3.5,"addressBlacklist":"","stopFollowTimes":0,"followSellPctMode":1,"onlyDevSwitch":1,"followClearPct":80,"stopFollowSwitch":1,"blockCount":30,"stopFollowSecond":1200,"quickSellSwitch":0,"quickSellMs":300,"quickSellPct":100,"mayhemSwitch":0,"drawdownPctSwitch":0,"drawdownActivateTarget":0,"drawdownPct":null},"akBuyDetails":[{"id":"6b3d673c-d3aa-4cff-bf63-fd5884305591","orderNum":1,"buyMode":1,"buyAmount":0.38,"slippageBuy":50,"marketValueMin":0,"marketValueMax":0,"gasFee":0.001,"tipFee":0.003,"highGasMode":1,"slippageMode":0,"highGasSlippage":50,"highGas":0.003,"highGasTip":0.002}],"akProfitTpSegments":[{"id":"aa77bc11-b6b5-4787-93e0-c6de7d39f96c","triggerPct":60,"sellPct":100,"timeOutSwitch":1,"timeOutStamp":2,"timeoutAtTriggerPct":100,"segmentOrder":1}],"akLossSegments":[{"id":"4736d4e7-c6d6-4026-b8c9-596dbde53b47","triggerPct":5,"sellPct":100,"segmentOrder":1}]}`, address)
	payload := strings.NewReader(payloadStr)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Authorization", "Bearer 8aa045e8464a4281aef37a17b4617935")
	req.Header.Add("apifoxtoken", "XL299LiMEDZ0H5h3A29PxwQXdMJqWyY2")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer res.Body.Close()

	// 检查 HTTP 状态码
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("API 返回错误状态码: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析 JSON 响应
	var resp AddAKAddressResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		// 如果解析失败，打印原始响应以便调试
		fmt.Printf("JSON 解析失败，原始响应: %s\n", string(body))
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 检查 API 返回的 code
	if resp.Code != 0 {
		return fmt.Errorf("API 返回错误: code=%d, message=%s", resp.Code, resp.Msg)
	}

	fmt.Printf("成功添加复制交易地址: %s, 响应: %d\n", address, resp.Data)
	return nil
}
func AddCopyAddress() {
	addresses, err := GetAddressTokenList()
	if err != nil {
		panic(err)
	}
	tradewizAddresses, err := GetTradewizCopyAddress()
	if err != nil {
		panic(err)
	}
	fmt.Println("跟单有:", len(tradewizAddresses), "个地址")
	// ak添加
	// for _, address := range tradewizAddresses {
	// 	time.Sleep(1000 * time.Millisecond)
	// 	err := AddAKAddress(address)
	// 	if err != nil {
	// 		fmt.Printf("添加失败: %s, 错误: %v\n", address, err)
	// 	} else {
	// 		fmt.Println("添加成功:", address)
	// 	}
	// }
	// return
	// ak
	for _, address := range addresses {
		// 慢点处理，避免请求过于频繁
		time.Sleep(500 * time.Millisecond)
		creator, err := GetTokenDetail(address)
		if err != nil {
			fmt.Printf("获取 token 详情失败: %s, 错误: %v\n", address, err)
			continue
		}
		if !slices.Contains(tradewizAddresses, creator) {
			err := AddTradewizCopyAddress(creator)
			if err != nil {
				fmt.Printf("添加失败: %s, 错误: %v\n", creator, err)
			} else {
				tradewizAddresses = append(tradewizAddresses, creator)
				fmt.Println("添加成功:", creator)
			}
		} else {
			fmt.Println("已存在:", creator)
		}
	}
}
