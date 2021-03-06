package message

import (
	"bytes"
	"fmt"
	"sort"
	"text/tabwriter"
)

// MakeReminderSendText Slackへリマインダーを送信する用のメッセージを作成する
func MakeReminderSendText(targetUserList map[string]int) string {
	count := len(targetUserList)
	if count == 0 {
		return `
<!channel>
今週は全員がブログを書きました！ :tada:
やったね！！！
`
	}
	return fmt.Sprintf(`
<!channel>
まだブログを書けていないユーザーがいます！
今週中に書けるようみんなで煽りましょう！
書けていないユーザー: %d人
================
%s`, count, getReminderReplaceMessageList(targetUserList))
}

// MakeResultSendText Slackへ先週の結果を送信するようのメッセージを作成する
func MakeResultSendText(maxBlogQuota int, targetUserList map[string]int) string {
	return fmt.Sprintf(`
<!channel>
1週間お疲れ様でした！
今週も頑張ってブログを書きましょう！
先週ブログを書けていない人は今週書くブログ記事が増えていることを確認してください！
================
%s================

%s
`,
		getReminderReplaceMessageList(filter(targetUserList, func(count int) bool {
			return count <= maxBlogQuota
		})),
		getCancelReplaceMessageList(filter(targetUserList, func(count int) bool {
			return count > maxBlogQuota
		})))
}

func filter(targetUserList map[string]int, judge func(int) bool) map[string]int {
	filteredUserList := make(map[string]int, len(targetUserList))
	for k, v := range targetUserList {
		if judge(v) {
			filteredUserList[k] = v
		}
	}
	return filteredUserList
}

// getCancelReplaceMessageList 退会処理用のユーザーリスト文字列を取得する
func getCancelReplaceMessageList(filteredUserList map[string]int) string {
	if len(filteredUserList) == 0 {
		return "今週は退会対象者はいません！ :tada:"
	}
	return fmt.Sprintf(`残念ながら以下の方は退会となります :cry:
================
%s================`, getReminderReplaceMessageList(filteredUserList))
}

// getReminderReplaceMessageList リマインダー用のユーザーリスト文字列を取得する
func getReminderReplaceMessageList(targetUserList map[string]int) string {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 4, 4, ' ', 0)
	names := make([]string, 0, len(targetUserList))
	for name := range targetUserList {
		names = append(names, name)
	}
	sort.Strings(names) //sort by key
	for _, n := range names {
		tw.Write([]byte(fmt.Sprintf("<@%s>さん\t残り%d記事\n", n, targetUserList[n])))
	}
	if err := tw.Flush(); err != nil {
		return fmt.Sprintf("リスト生成に失敗 %+v\n", targetUserList)
	}

	return buf.String()
}
