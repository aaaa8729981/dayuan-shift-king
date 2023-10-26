package main

import (
  "fmt"
  "log"
  "net/http"
  "strings"
  "time"

  "github.com/line/line-bot-sdk-go/v7/linebot"
  "os"
  "time"
  "strconv"
)

// 定義一個全局變量用於記錄上次觸發sumall的時間
var lastSumAllTriggerTime time.Time

func main() {  
  // 定義：透過groupID取得指定群組成員列表(userID)
  groupID := os.Getenv("LINEBOTGROUP_ID")
  memberIDs, err := getGroupUserIDs(groupID)
  if err != nil {
      log.Println("取得群組成員列表失败:", err)
  }

  // 从环境变量获取时间设置值
  workMessageHour1, err := strconv.Atoi(os.Getenv("WORKMESSAGEHOUR1"))
  if err != nil {
      log.Println("无法解析WORKMESSAGEHOUR1环境变量", err)
      workMessageHour1 = 11
    }
  workMessageMinute1, err := strconv.Atoi(os.Getenv("WORKMESSAGEMINUTE1"))
  if err != nil {
      log.Println("无法解析WORKMESSAGEMINUTE1环境变量", err)
      workMessageMinute1 = 0
  }

  workMessageHour2, err := strconv.Atoi(os.Getenv("WORKMESSAGEHOUR2"))
  if err != nil {
      log.Println("无法解析WORKMESSAGEHOUR2环境变量", err)
      workMessageHour2 = 20
  }

  workMessageMinute2, err := strconv.Atoi(os.Getenv("WORKMESSAGEMINUTE2"))
  if err != nil {
      log.Println("无法解析WORKMESSAGEMINUTE2环境变量", err)
      workMessageMinute2 = 30
  }

  // 调用其他函数，并传递环境变量的值作为参数
  triggerWorkMessage(workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2)
  
  
  // 定時觸發 SumAll
  go triggerSumAllDaily()
  // 定时触发 "上班囉" 消息
  go triggerWorkMessage()
}

// 取得指定群組成員的用戶id(userID)
func getGroupUserIDs(groupID string) ([]string, error) {
    continuationToken := ""
    var memberIDs []string

    for {
        resp, err := bot.GetGroupMemberIDs(groupID, continuationToken).Do()
        if err != nil {
            return nil, err
        }

        memberIDs = append(memberIDs, resp.MemberIDs...)
        continuationToken = resp.Next
        if continuationToken == "" {
            break
        }
    }

    return memberIDs, nil
}

  // 函数用于计算等待的时间
  func calculateWaitTime(targetTime time.Time) time.Duration {
    now := time.Now()
    if now.After(targetTime) {
      targetTime = targetTime.Add(24 * time.Hour)
    }
    return targetTime.Sub(now)
  }

  // 函数用于发送消息
  func sendMessage(bot *linebot.Client, groupID string, message string) error {
    _, err := bot.PushMessage(groupID, linebot.NewTextMessage(message)).Do()
    return err
  }

  
      func triggerWorkMessage(workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2 int) {
        for {
            now := time.Now()
            weekday := now.Weekday()

        // 僅在星期一到星期五执行
        if weekday >= time.Monday && weekday <= time.Friday {
          targetTime1 := time.Date(now.Year(), now.Month(), now.Day(), workMessageHour1, workMessageMinute1, 0, 0, time.Local)
          targetTime2 := time.Date(now.Year(), now.Month(), now.Day(), workMessageHour2, workMessageMinute2, 0, 0, time.Local)

          timeToWait1 := calculateWaitTime(targetTime1)
          timeToWait2 := calculateWaitTime(targetTime2)

          // 选择等待时间较短的时间来触发消息
          var timeToWait time.Duration
          if timeToWait1 < timeToWait2 {
              timeToWait = timeToWait1
          } else {
              timeToWait = timeToWait2
          }

          // 等待时间后触发消息
          <-time.After(timeToWait)
          sendMessage(bot, groupID, "上班囉")

          // 设置定时触发 SumAll 的计时器
          go triggerSumAll(event)
      }
    }
  }



// 觸發sumall(在發送pushMessage之後的30分鐘)
func triggerSumAll(event *linebot.Event) {
  count, err := strconv.Atoi(os.Getenv("SUMALLTRIGGERCOUNT"))
  if err != nil {
    log.Println("無法解析SUMALLTRIGGERCOUNT環境變量", err)
    return
  }

  for i := 0; i < count; i++ {
    time.Sleep(30 * time.Minute)

    // 触发 SumAll
    handleSumAll(event)

    // 更新上次触发 SumAll 的时间
    lastSumAllTriggerTime = time.Now()
  }
}


func callbackHandler(w http.ResponseWriter, r *http.Request) {
  events, err := bot.ParseRequest(r)

  if err != nil {
    if err == linebot.ErrInvalidSignature {
      w.WriteHeader(http.StatusBadRequest)
    } else {
      w.WriteHeader(http.StatusInternalServerError)
    }
    return
  }

  for _, event := range events {
    if event.Type == linebot.EventTypeMessage {
      switch message := event.Message.(type) {
      // Handle only on text message
      case *linebot.TextMessage:
        // Directly to ChatGPT
        if strings.Contains(message.Text, ":gpt") {
          // New feature.
          if IsRedemptionEnabled() {
            if stickerRedeemable {
              handleGPT(GPT_Complete, event, message.Text)
              stickerRedeemable = false
            } else {
              handleRedeemRequestMsg(event)
            }
          } else {
            // Original one
            handleGPT(GPT_Complete, event, message.Text)
          }
        } else if strings.Contains(message.Text, ":gpt4") {
          // New feature.
          if IsRedemptionEnabled() {
            if stickerRedeemable {
              handleGPT(GPT_GPT4_Complete, event, message.Text)
              stickerRedeemable = false
            } else {
              handleRedeemRequestMsg(event)
            }
          } else {
            // Original one
            handleGPT(GPT_GPT4_Complete, event, message.Text)
          }
        } else if strings.Contains(message.Text, ":draw") {
          // New feature.
          if IsRedemptionEnabled() {
            if stickerRedeemable {
              handleGPT(GPT_Draw, event, message.Text)
              stickerRedeemable = false
            } else {
              handleRedeemRequestMsg(event)
            }
          } else {
            // Original one
            handleGPT(GPT_Draw, event, message.Text)
          }
        } else if strings.EqualFold(message.Text, ":list_all") && isGroupEvent(event) {
          handleListAll(event)
        } else if strings.EqualFold(message.Text, "sum_all") && isGroupEvent(event) {
          handleSumAll(event)
        } else if isGroupEvent(event) {
          // 如果聊天機器人在群組中，開始儲存訊息。
          handleStoreMsg(event, message.Text)
        }

      // Handle only on Sticker message
      case *linebot.StickerMessage:
        var kw string
        for _, k := range message.Keywords {
          kw = kw + "," + k
        }

        log.Println("Sticker: PID=", message.PackageID, " SID=", message.StickerID)
        if IsRedemptionEnabled() {
          if message.PackageID == RedeemStickerPID && message.StickerID == RedeemStickerSID {
            stickerRedeemable = true
            if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("你的賦能功能啟動了！")).Do(); err != nil {
              log.Print(err)
            }
          }
        }

        if isGroupEvent(event) {
          // 在群組中，一樣紀錄起來不回覆。
          outStickerResult := fmt.Sprintf("貼圖訊息: %s ", kw)
          handleStoreMsg(event, outStickerResult)
        } else {
          outStickerResult := fmt.Sprintf("貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerID, message.PackageID, kw, message.Text)

          // 1 on 1 就回覆
          if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outStickerResult)).Do(); err != nil {
            log.Print(err)
          }
        }
      }
    }
  }
}

func handleSumAll(event *linebot.Event) {
  // Scroll through all the messages in the chat group (in chronological order).
  oriContext := ""
  q := summaryQueue.ReadGroupInfo(getGroupID(event))
  for _, m := range q {
    // [xxx]: 他講了什麼... 時間
    oriContext = oriContext + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, m.Time.Local().UTC().Format("2006-01-02 15:04:05"))
  }

  // 取得使用者暱稱
  userName := event.Source.UserID
  // 取得群組ID
  groupID := event.Source.GroupID
  // 取得指定群组成員的個人資料，在這裡的userName其實是webhook裡面的event.Source.UserID（一次只能查一個）
  profile, err := bot.GetGroupMemberProfile(groupID, userName).Do()
  if err != nil {
      // 處理錯誤
      log.Println("取得指定群組成員個人資料錯誤:", err)
  } else {
      // 使用 profile 中的信息，例如 profile.DisplayName
      userName := profile.DisplayName
  }

  // 訊息內先回，再來總結。
  // if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("好的，總結文字已經發給您了"+userName)).Do(); err != nil {
  // 	log.Print(err)
  // }

  // 就是請 ChatGPT 幫你總結
  // 尚未完成：在oriContext中放入群組清單
  oriContext = fmt.Sprintf("下面的许多讯息是一个工作的群组，请将以下内容统整，原则上依照内容里的时间排序。請用繁體中文回覆，如果內容無法理解，不需統整沒關係，直接列出即可。請不要捏造內容\n成员: %s\n\n%s", userName, oriContext)
  reply := gptGPT3CompleteContext(oriContext)

  // 在群組中使用ReplyToken回覆訊息
  if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前總結如下：\n" + reply)).Do(); err != nil {
    log.Print(err)
  }
}

func handleListAll(event *linebot.Event) {
  reply := ""
  q := summaryQueue.ReadGroupInfo(getGroupID(event))
  for _, m := range q {
    reply = reply + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, m.Time.Local().UTC().Format("2006-01-02 15:04:05"))
  }

  if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
    log.Print(err)
  }
}

func handleGPT(action GPT_ACTIONS, event *linebot.Event, message string) {
  switch action {
  case GPT_Complete:
    reply := gptGPT3CompleteContext(message)
    if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
      log.Print(err)
    }
  case GPT_GPT4_Complete:
    reply := gptGPT4CompleteContext(message)
    if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
      log.Print(err)
    }
  case GPT_Draw:
    if reply, err := gptImageCreate(message); err != nil {
      if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("無法正確顯示圖形.")).Do(); err != nil {
        log.Print(err)
      }
    } else {
      if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("根據你的提示，畫出以下圖片："), linebot.NewImageMessage(reply, reply)).Do(); err != nil {
        log.Print(err)
      }
    }
  }

}

func handleRedeemRequestMsg(event *linebot.Event) {
  // First, obtain the user's Display Name (i.e., the name displayed).
  userName := event.Source.UserID
  userProfile, err := bot.GetProfile(event.Source.UserID).Do()
  if err == nil {
    userName = userProfile.DisplayName
  }

  if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(userName+":你需要買貼圖，開啟這個功能"), linebot.NewStickerMessage(RedeemStickerPID, RedeemStickerSID)).Do(); err != nil {
    log.Print(err)
  }
}

func handleStoreMsg(event *linebot.Event, message string) {
  // Get user display name. (It is nick name of the user define.)
  userName := event.Source.UserID
  userProfile, err := bot.GetProfile(event.Source.UserID).Do()
  if err == nil {
    userName = userProfile.DisplayName
  }

  // event.Source.GroupID 就是聊天群組的 ID，並且透過聊天群組的 ID 來放入 Map 之中。
  m := MsgDetail{
    MsgText:  message,
    UserName: userName,
    Time:     time.Now(),
  }
  summaryQueue.AppendGroupInfo(getGroupID(event), m)
}

func isGroupEvent(event *linebot.Event) bool {
  return event.Source.GroupID != "" || event.Source.RoomID != ""
}

func getGroupID(event *linebot.Event) string {
  if event.Source.GroupID != "" {
    return event.Source.GroupID
  } else if event.Source.RoomID != "" {
    return event.Source.RoomID
  }

  return ""
}
