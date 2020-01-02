package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"net/http"
	"time"
)

const (
	SlackApi   string = ""
	DateLayout string = "2006-01-02"
)

func NewCostExplorerClient() *costexplorer.CostExplorer {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("ap-northeast-1"),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))
	return costexplorer.New(sess, aws.NewConfig().WithRegion("ap-northeast-1"))
}

type CostInfo struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Amount string `json:"amount"`
}

type SlackMessage struct {
	Text string `json:"text"`
	Mrkdwn bool `json:"mrkdwn"`
}

type TimeStringHelper struct {
	Location *time.Location
	Now time.Time
}

func NewTimeStringHelper(locationName string) *TimeStringHelper {
	location, _ := time.LoadLocation(locationName)
	return &TimeStringHelper{
		Location: location,
		Now: time.Now().In(location),
	}
}

func (helper TimeStringHelper) GetBeginningOfLastMonth() string {
	return time.Date(helper.Now.Year(), helper.Now.Month() -1, 1, 0, 0, 0, 0, helper.Location).Format(DateLayout)
}

func (helper TimeStringHelper) GetBeginningOfMonth() string {
	return time.Date(helper.Now.Year(), helper.Now.Month(), 1, 0, 0, 0, 0, helper.Location).Format(DateLayout)
}

func (helper TimeStringHelper) GetYesterday() string {
	return time.Date(helper.Now.Year(), helper.Now.Month(), helper.Now.Day() -1, 0, 0, 0, 0, helper.Location).Format(DateLayout)
}

func (helper TimeStringHelper) GetToday() string {
	return time.Date(helper.Now.Year(), helper.Now.Month(), helper.Now.Day(), 0, 0, 0, 0, helper.Location).Format(DateLayout)
}

func (helper TimeStringHelper) IsTodayFirst() bool {
	return helper.Now.Day() == 1
}

func (helper TimeStringHelper) GetStartTimePeriod() string {
	if helper.IsTodayFirst() {
		return helper.GetBeginningOfLastMonth()
	} else {
		return helper.GetBeginningOfMonth()
	}
}


func GetCostInfo(helper *TimeStringHelper) *CostInfo {
	start := helper.GetStartTimePeriod()
	end := helper.GetToday()
	costExplorer := NewCostExplorerClient()
	output, err := costExplorer.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
		Granularity: aws.String("MONTHLY"),
		Metrics: []*string{
			aws.String("AmortizedCost"),
		},
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
	})
	if err != nil {
		panic(err)
	}
	total := output.ResultsByTime[0].Total["AmortizedCost"]
	amount := aws.StringValue(total.Amount)

	return &CostInfo{
		Start:  start,
		End:    end,
		Amount: amount,
	}
}

func makeSlackMessage(costInfo *CostInfo, helper *TimeStringHelper) SlackMessage {
	return SlackMessage{
		Text: fmt.Sprintf("*期間*: `%s ~ %s`\n*料金*: `$%s`",
			costInfo.Start,
			helper.GetYesterday(),
			costInfo.Amount),
		Mrkdwn: true}
}

func PostToSlack(message SlackMessage) {
	input, _ := json.Marshal(message)
	fmt.Println(string(input))
	http.Post(SlackApi, "application/json", bytes.NewBuffer(input))
}

type Response struct {
	Message []byte `json:"message"`
}

func BillingNotification(ctx context.Context) (Response, error) {
	helper := NewTimeStringHelper("Asia/Tokyo")
	fmt.Println(helper.Now.Format("2006/01/02 15:04:05"))
	costInfo := GetCostInfo(helper)
	message := makeSlackMessage(costInfo, helper)
	PostToSlack(message)
	json, _ := json.Marshal(message)
	return Response{Message: json}, nil
}

func main() {
	lambda.Start(BillingNotification)
}
