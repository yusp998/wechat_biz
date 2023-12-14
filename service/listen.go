package service

import (
	"github.com/gin-gonic/gin"
	"log"
)

type ListenBody struct {
	MsgType      int         `json:"msgtype"`
	Wxid         string      `json:"wxid"`
	content      string      `json:"content"`
	RoomName     string      `json:"RoomName"`
	sender       string      `json:"sender"`
	SendNickName string      `json:"SendNickName"`
	FilePath     string      `json:"filepath"`
	SelfWxid     string      `json:"selfwxid"`
	Atuserlist   interface{} `json:"atuserlist"`
	IsPhoneMsg   int         `json:"isphonemsg"`
	Timestamp    int         `json:"timestamp"`
	Port         int         `json:"port"`
}

func ListenAllMsg(c *gin.Context) {
	//var listenBody = ListenBody{}
	//c.ShouldBindJSON(&listenBody)
	//b, _ := json.Marshal(listenBody)
	//log.Println(string(b))

	json := make(map[string]interface{}) //注意该结构接受的内容
	c.BindJSON(&json)
	log.Printf("%v", &json)
}
