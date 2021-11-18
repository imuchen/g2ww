package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber"
)

type Hook struct {
	DashboardId int         `json:"dashboardId"`
	EvalMatches []evalMatch `json:"evalMatches"`
	ImageUrl    string      `json:"imageUrl"`
	// Alert message
	Message  string `json:"message"`
	OrgId    int    `json:"orgId"`
	PanelId  int    `json:"panelId"`
	RuleId   int    `json:"ruleId"`
	RuleName string `json:"ruleName"`
	RuleUrl  string `json:"ruleUrl"`
	State    string `json:"state"`
	tags     string `json:"tags"`
	// Panel Title
	Title string `json:"title"`
}

type evalMatch struct {
	Metric string  `json:"metric"`
	Value  float64 `json:"value"`
	tags   string  `json:"tags"`
}

var sent_count int = 0

func GwStat() func(c *fiber.Ctx) {
	return func(c *fiber.Ctx) {
		stat_msg := "G2WW Server created by Nova Kwok is running! \nParsed & forwarded " + strconv.Itoa(sent_count) + " messages to WeChat Work!"
		c.Send(stat_msg)
		return
	}
}

func processMatches(evalMatches []evalMatch) string {
	//var result string = "\n告警明细：\nMetric|Value"
	var result string = ""
	var j int
	for j = 0; j < len(evalMatches); j++ {
		result += evalMatches[j].Metric + ":" + strconv.FormatFloat(evalMatches[j].Value, 'f', 6, 64)
	}
	result = strings.Replace(result, "\"", "\\\"", -1)
	return result
}

func GwWorker() func(c *fiber.Ctx) {
	return func(c *fiber.Ctx) {
		h := new(Hook)
		if err := c.BodyParser(h); err != nil {
			fmt.Println(err)
			c.Send("Error on JSON format")
			return
		}

		url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + c.Params("key")

		msgStr := fmt.Sprintf(`
		{
			"msgtype": "news",
			"news": {
			  "articles": [
				{
				  "title": "%s",
				  "description": "%s",
				  "url": "%s",
				  "picurl": "%s"
				}
			  ]
			}
		  }
		`, h.Title, h.Message+processMatches(h.EvalMatches), strings.Replace(h.RuleUrl, "editPanel=", "viewPanel=", -1), h.ImageUrl)
		fmt.Println(time.Now().Format("2006-01-02 15:04:05") + msgStr)
		jsonStr := []byte(msgStr)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.Send("Error sending to WeChat Work API")
			fmt.Println(time.Now().Format("2006-01-02 15:04:05") + err.Error())
			return
		}
		defer resp.Body.Close()
		c.Send(resp)
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println(resp)
		fmt.Println("==============================================")
		sent_count++
	}
}
