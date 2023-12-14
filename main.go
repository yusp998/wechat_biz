package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"wechat_article_spider/initdata"
	"wechat_article_spider/service"
)

func init() {
	err := initdata.InitMySQLCon()
	if err != nil {
		log.Printf("initdata sql err %s", err.Error())
		return
	}
	initdata.InitRedis()
}

func main() {
	router := gin.Default()
	v1 := router.Group("/wechat/article/v1/")
	{
		v1.POST("/attentionBiz", service.AttentionBiz)
		v1.POST("/bizList", service.BizList)
		v1.POST("/articleListCrawler", service.ArticleListCrawler)
		v1.POST("/articleContentCrawler", service.ContentCrawler)
		v1.POST("/articleNumsCrawler", service.NumsCrawler)
	}
	//router.POST("/msg", service.ListenAllMsg)
	router.Run(":8888")
}
