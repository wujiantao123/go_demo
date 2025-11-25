package pumpfollow

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Following struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Address  string `json:"address"`
	Name     string `json:"name"`
	Remark   string `json:"remark"`
}

func GetPumpFollowings() {
	data, err := os.ReadFile("pkg/remark.json")
	if err != nil {
		panic(err)
	}

	var followings []Following
	if err := json.Unmarshal(data, &followings); err != nil {
		log.Fatalf("无法解析JSON: %v", err)
	}

	for _, following := range followings {
		fmt.Printf("%s,%s \n", following.Address, following.Remark)
	}
}
