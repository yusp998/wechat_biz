package service

import (
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"time"
	"wechat_article_spider/api"
	"wechat_article_spider/model"
)

func NumsCrawler(c *gin.Context) {
	var body struct {
		BizName string `json:"bizName" binding:"required"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ResponseData{Message: "参数错误", Code: 400})
		return
	}
	go doNumsCrawler(body.BizName)
	c.JSON(http.StatusOK, ResponseData{Message: "suc", Code: 200})
}

func doNumsCrawler(bizName string) {
	articles := model.SelectUrlByBizName(bizName)
	for _, v := range articles {
		log.Printf("bizName %s url %s", bizName, v.Url)
		jsonStr := api.ArticleNums(v.Url)
		if jsonStr != "" && gjson.Valid(jsonStr) {
			readNum := gjson.Get(jsonStr, "appmsgstat.read_num").Int()
			likeNum := gjson.Get(jsonStr, "appmsgstat.old_like_num").Int()
			lookNum := gjson.Get(jsonStr, "appmsgstat.like_num").Int()
			err := model.UpdateNums(v.Id, readNum, likeNum, lookNum)
			if err != nil {
				log.Printf("sql UpdateNums err readum %d likeNum %d lookNum %d url %s", readNum, likeNum, lookNum, v.Url)
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	log.Printf("bizName %s numsCrawler end", bizName)
}
