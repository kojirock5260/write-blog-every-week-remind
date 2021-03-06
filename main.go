package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/write-blog-every-week/write-blog-every-week-remind/config"
	"github.com/write-blog-every-week/write-blog-every-week-remind/database"
	"github.com/write-blog-every-week/write-blog-every-week-remind/date"
	"github.com/write-blog-every-week/write-blog-every-week-remind/message"
	"github.com/write-blog-every-week/write-blog-every-week-remind/rss"
	"github.com/write-blog-every-week/write-blog-every-week-remind/slack"
)

// main
func main() {
	executeFunction := os.Getenv("GOLANG_EXECUTE_FUNCTION")
	if executeFunction == "reminder" {
		lambda.Start(blogReminder)
		// blogReminder()
	} else if executeFunction == "register" {
		lambda.Start(blogRegister)
	} else if executeFunction == "result" {
		lambda.Start(blogResult)
		// blogResult()
	} else {
		panic("環境変数 GOLANG_EXECUTE_FUNCTION が取得出来ないか、期待した値ではありません。")
	}
}

// blogReminder ブログのリマインダーロジックを実行
func blogReminder() {
	thisMonday := date.GetThisMonday()
	configData := config.GetConfigData()
	allMemberDataList := database.FindAll(configData)
	targetUserList, _ := rss.FindTargetUserList(allMemberDataList, thisMonday)

	for u, c := range targetUserList {
		if c == 0 {
			delete(targetUserList, u)
		}
	}

	sendText := message.MakeReminderSendText(targetUserList)
	slack.SendMessage(configData, sendText)
	// fmt.Println(sendText)
}

// blogRegister ブログの登録ロジックを実行
func blogRegister(_ context.Context, rawParams interface{}) (interface{}, error) {
	configData := config.GetConfigData()
	envToken := os.Getenv("WBEW_SLACK_TOKEN")
	params, err := slack.ParseSlackParams(rawParams)
	if err != nil {
		return "スラックのパラメータが取得できませんでした。 error: " + err.Error(), nil
	}
	if envToken != params.Token {
		return "トークンの不一致", nil
	}
	userData := database.FindByPK(configData, params.UserID)
	if userData.UserID != "" {
		return "あなたのブログはすでに登録済みです feedURL: " + userData.FeedURL, nil
	}
	database.CreateUser(configData, params)

	return "ブログを登録しました。これからは妥協は許しませんよ。", nil
}

// blogResult ブログ書けたかどうか通知のロジックを実行
func blogResult() {
	lastWeekMonday := date.GetLastWeekMonday()
	configData := config.GetConfigData()
	allMemberDataList := database.FindAll(configData)
	targetUserList, _ := rss.FindTargetUserList(allMemberDataList, lastWeekMonday)

	for userID := range targetUserList {
		// 0の人は1になり、1以上の人は1記事増える
		targetUserList[userID]++
	}

	database.ResetRequireCount(configData, targetUserList)
	sendText := message.MakeResultSendText(configData.Blog.MaxBlogQuota, targetUserList)
	slack.SendMessage(configData, sendText)
	// fmt.Println(sendText)
}
