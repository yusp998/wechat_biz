package service

import (
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wechat_article_spider/api"
	"wechat_article_spider/initdata"
	"wechat_article_spider/model"
	"wechat_article_spider/util"
)

type ResponseData struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Code    int64       `json:"code"`
}

var articleListRedisPrefix = "wechat_go_list_"

func AttentionBiz(c *gin.Context) {
	var body struct {
		BizName  string `json:"bizName" binding:"required"`
		BizTitle string `json:"bizTitle" default:""`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ResponseData{Message: "参数错误", Code: 400})
		return
	}
	if body.BizName != "" {
		bizInfo := model.SelectByBizName(body.BizName)
		if bizInfo.Id > 0 {
			c.JSON(http.StatusOK, ResponseData{Message: "已关注", Code: 200})
			return
		}
		str := api.AttentionBiz(body.BizName)
		if strings.Contains(str, "请求已经被成功处理") {
			time.Sleep(3 * time.Second)
			jsonStr := api.SelectByBiz(body.BizName)
			title := body.BizTitle
			bizName := body.BizName
			if jsonStr != "" && gjson.Valid(jsonStr) {
				array := gjson.Get(jsonStr, "data").Array()
				if len(array) > 0 {
					info := array[0]
					title = info.Get("Title").Str

				}
			}
			biz := model.Biz{
				BizTitle: title,
				BizName:  bizName,
				Status:   1,
			}
			if biz.Insert() == nil {
				c.JSON(http.StatusOK, ResponseData{Message: "suc", Code: 200})
				return
			}
		}
	}
	c.JSON(http.StatusBadGateway, ResponseData{Message: "未知错误", Code: 500})

}

func BizList(c *gin.Context) {
	var body struct {
		BizTitle string `json:"bizTitle" default:""`
	}
	c.ShouldBindJSON(&body)
	bizs := model.SelectByBizTitle(body.BizTitle)
	c.JSON(http.StatusOK, ResponseData{Message: "suc", Code: 200, Data: bizs})
}

func ArticleListCrawler(c *gin.Context) {
	var body struct {
		BizName string `json:"bizName"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ResponseData{Message: "参数错误", Code: 400})
		return
	}

	n, err := initdata.Redisdb.Exists(initdata.CTX, articleListRedisPrefix+body.BizName).Result()
	if err != nil {
		c.JSON(http.StatusBadGateway, ResponseData{Message: "redis error", Code: 500})
		return
	}
	if n > 0 {
		c.JSON(http.StatusOK, ResponseData{Message: "正在执行中请稍后", Code: 200})
		return
	}
	go doArticleListCrawler(body.BizName)
	c.JSON(http.StatusOK, ResponseData{Message: "suc", Code: 200})
}

func doArticleListCrawler(bizName string) {
	initdata.Redisdb.Set(initdata.CTX, articleListRedisPrefix+bizName, "1", 24*time.Hour)
	articles := model.SelectUidByBizName(bizName)
	tempMap := make(map[string]interface{})
	lastUid := ""
	if len(articles) > 0 {
		lastUid = articles[0].UniqueId
		for _, v := range articles {
			tempMap[v.UniqueId] = ""
		}
	}
	log.Printf("获取已存在文章列表 size %d lastuid %s\n", len(articles), lastUid)

	offset := ""
	isEnd := 0
	for isEnd == 0 {
		jsonStr := api.ArticleListByWxid(bizName, offset, isEnd)
		time.Sleep(2 * time.Second)
		if jsonStr != "" && gjson.Valid(jsonStr) {
			msgArray := gjson.Get(jsonStr, "MsgList.Msg").Array()
			offset = gjson.Get(jsonStr, "MsgList.PagingInfo.Offset").Str
			isEnd, _ = strconv.Atoi(gjson.Get(jsonStr, "MsgList.PagingInfo.IsEnd").Raw)
			if len(msgArray) > 0 {
				log.Printf("获取微信文章列表 offset %s size %d", offset, len(msgArray))
				for _, info := range msgArray {
					publishTimeStr := info.Get("AppMsg.BaseInfo.UpdateTime").Raw
					publishTime, _ := strconv.ParseInt(publishTimeStr, 10, 64)
					detailArray := info.Get("AppMsg.DetailInfo").Array()
					baseUid := bizName + "_" + info.Get("BaseInfo.UniqueId").Str
					run := true
					if len(detailArray) > 0 && run {
						for _, detailInfo := range detailArray {
							url := detailInfo.Get("ContentUrl").Str
							idx := util.GetIDXByUrl(url)
							if idx == "" {
								log.Println("##########################")
								log.Printf("idx error %s\n", url)
								log.Println("##########################")
								continue
							}
							uniqueId := baseUid + "_" + idx
							//if uniqueId == lastUid {
							//	//fmt.Println("已获取到最新", uniqueId)
							//	log.Printf("已获取到最新 %s\n", uniqueId)
							//	//isEnd = 1
							//	run = false
							//	break
							//}
							if _, ok := tempMap[uniqueId]; ok {
								//fmt.Println("已存在文章", uniqueId)
								log.Printf("已存在文章 %s\n", uniqueId)
								continue
							}

							articleDesc := detailInfo.Get("Digest").Str
							title := detailInfo.Get("Title").Str
							aboutDesc := detailInfo.Get("ShowDesc").Str
							showImg := detailInfo.Get("SuggestedCoverImg.url").Str
							allImg := ""
							for k, v := range detailInfo.Map() {
								if strings.HasPrefix(k, "CoverImgUrl") {
									allImg += v.Str + ","
								}
							}
							article := model.Article{
								BizName:     bizName,
								UniqueId:    uniqueId,
								Title:       title,
								ArticleDesc: articleDesc,
								ShowImg:     showImg,
								AllImg:      allImg,
								PublishTime: publishTime,
								AboutDesc:   aboutDesc,
								Url:         url,
							}

							if aboutDesc != "" {
								aboutDesc = strings.Replace(aboutDesc, " ", "", -1)
								aboutDesc = strings.Replace(aboutDesc, " ", "", -1)
								aboutDesc = strings.Replace(aboutDesc, " ", "", -1)
								article.AboutReadNum = util.RegexpMatch(aboutDesc, "阅读(\\d+\\.\\d+|\\d+)")
								article.AboutLikeNum = util.RegexpMatch(aboutDesc, "赞(\\d+\\.\\d+|\\d+)")
								article.AboutLookNum = util.RegexpMatch(aboutDesc, "看过(\\d+\\.\\d+|\\d+)")
								article.AboutLookNum = util.RegexpMatch(aboutDesc, "观看(\\d+\\.\\d+|\\d+)")
								article.AboutDesc = aboutDesc
							}

							article.Insert()
						}
						if !run {
							isEnd = 1
							break
						}

					}

				}
			}
		}
	}
	log.Printf("bizName %s crawler arctile list end\n", bizName)
	initdata.Redisdb.Del(initdata.CTX, articleListRedisPrefix+bizName)
}
