package weixin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

/*
[{
	"scopeId": 2,
	"name": "growing-segmentation-pid:15149@seg3",
	"id0": 47,
	"id1": 0,
	"alarmMessage": "Response time of service instance growing-segmentation-pid:15149@seg3 is more than 1000ms in 2 minutes of last 10 minutes",
	"startTime": 1568888544862
}, {
	"scopeId": 2,
	"name": "growing-segmentation-pid:11847@seg2",
	"id0": 46,
	"id1": 0,
	"alarmMessage": "Response time of service instance growing-segmentation-pid:11847@seg2 is more than 1000ms in 2 minutes of last 10 minutes",
	"startTime": 1568888544862
}]
*/
type message struct {
	ScopeId      int
	Name         string
	Id0          int
	Id1          int
	AlarmMessage string
	StartTime    int64
}

// Weixin 发送企业微信消息体
func Weixin(data []byte) error {
	var m []message
	err := json.Unmarshal(data, &m)
	if err != nil {
		fmt.Println(err.Error())
	}
	contents, alertSummary := createContent(m)
	bodys := strings.NewReader(contents)
	url := viper.GetString("weixin.url")
	resp, err := http.Post(url, "application/json", bodys)
	if err != nil {
		return err
	}
	log.Println(resp.StatusCode, alertSummary)
	return nil
}

/*
状态: notify

等级: P1

告警: Skywalking
  growing-segmentation-pid:6494@seg1  id: 44  time: 1568945304861
  growing-segmentation-pid:6908@seg0  id: 43  time: 1568945304861


Item values:

0  Response time of service instance growing-segmentation-pid:6494@seg1 is more than 1000ms in 2 minutes of last 10 minutes
1  Response time of service instance growing-segmentation-pid:6908@seg0 is more than 1000ms in 2 minutes of last 10 minutes


故障修复:
*/
func createContent(message []message) (string, string) {
	var alertname bytes.Buffer
	var alertSummary bytes.Buffer

	for i, alert := range message {
		if i > 0 {
			alertname.WriteString("\n\n")
		}

		timestamp := formatTime(alert.StartTime)
		alertname.WriteString(fmt.Sprintf("应用名: %s\n报警时间: %s\n报警号: %d\n报警内容: %s", alert.Name, timestamp, alert.Id0, alert.AlarmMessage))
		alertSummary.WriteString(fmt.Sprintf("%b,%s\n", i, alert.AlarmMessage))
	}

	contents := fmt.Sprintf("Skywalking告警: \n%s", alertname.String())
	data := fmt.Sprintf(`{
        "msgtype": "text",
            "text": {
            "content": "%s",
        }
    }`, contents)
	return data, alertSummary.String()
}

func formatTime(timestamp int64) string {
	tm := time.Unix(timestamp/1000, 0)
	return tm.Format("2006-01-02 15:04:05")
}
