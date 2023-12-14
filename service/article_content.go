package service

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
	"wechat_article_spider/model"
	"wechat_article_spider/util"
)

func ContentCrawler(c *gin.Context) {
	var body struct {
		BizName string `json:"bizName" binding:"required"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ResponseData{Message: "参数错误", Code: 400})
		return
	}
	go doContentCrawler(body.BizName)
	c.JSON(http.StatusOK, ResponseData{Message: "suc", Code: 200})
}

func doContentCrawler(bizName string) {
	articles := model.SelectUrlByBizName(bizName)
	for _, v := range articles {
		log.Printf("bizName %s content crawler url %s", bizName, v.Url)
		content, _ := util.AutoGetArticleContent(v.Url)
		if content != "" {
			err := model.UpdateContent(v.Id, content)
			if err != nil {
				println(err)
			}
		}
		time.Sleep(2 * time.Second)
	}
	log.Printf("bizName %s content crawler end", bizName)
}
