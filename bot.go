package main

import (
  "fmt"
  "log"
  "net/http"
  "strings"
  "time"
  "github.com/line/line-bot-sdk-go/v7/linebot"
  "os"
  "strconv"
)

// 定義一個全局變量用於記錄上次觸發sumall的時間
var lastSumAllTriggerTime time.Time
var groupMemberProfile string // 將 groupMemberProfile 變數宣告為全局變數
var taipeiLocation *time.Location？ //全局變量：將時區指定為台北


func initializeGroup() (string, []string, string, int, int, int, int, *time.Location) {
  if messageSent {
    // 消息已经发送，不需要再发送
    return "", nil, "", 0, 0, 0, 0
  }

  // 从 env 中获取 LINEBOTGROUP_ID
  groupIDFromEnv := os.Getenv("LINEBOTGROUP_ID")

  var groupID string
  if groupIDFromEnv != "" {
      groupID = groupIDFromEnv //groupID是從env裡面設定的
  } else {
    // 如果环境变量中群组 ID 为空，记录日志并跳过这个功能
    log.Println("未设置 Env 的 groupID")
    return "", nil, "", 0, 0, 0, 0
  }

  // 定义：透过 groupID 取得指定群组成员列表 (userID)
  memberIDsResponse, err := bot.GetGroupMemberIDs(groupID, "").Do()
  var userNames []string

  if err != nil {
      log.Println("透過env取得群組成員列表:") //取得失敗時，不判定為錯誤。userNames繼續維持為空值
  } else {
      for _, userID := range memberIDsResponse.MemberIDs {
          userNames = append(userNames, userID)
      }
      log.Println("群组成员 id 列表:", userNames)
  }

  groupMemberProfile = ""
  if len(userNames) > 0 { //如果userNames長度>0表示有成功取得如果userNames
    for _, userName := range userNames {
      profile, err := bot.GetGroupMemberProfile(groupID, userName).Do()
      if err != nil {
          log.Printf("獲取群組成員名稱資料錯誤 (用户名: %s): %v", userName, err)
      } else {
          groupMemberProfile += profile.DisplayName + ","
            }
        }
    } else {
    // userNames 为空，从环境变量中获取 groupMemberProfile
  groupMemberProfile = os.Getenv("GROUPMEMBERPROFILE")
  }
  groupMemberProfile = strings.TrimSuffix(groupMemberProfile, ",")

  // 从环境变量获取时间设置值
  workMessageHour1, err := strconv.Atoi(os.Getenv("WORKMESSAGEHOUR1"))
  if err != nil {
      log.Println("无法解析 WORKMESSAGEHOUR1 环境变量", err)
      workMessageHour1 = 11
  }
  workMessageMinute1, err := strconv.Atoi(os.Getenv("WORKMESSAGEMINUTE1"))
  if err != nil {
      log.Println("无法解析 WORKMESSAGEMINUTE1 环境变量", err)
      workMessageMinute1 = 0
  }

  workMessageHour2, err := strconv.Atoi(os.Getenv("WORKMESSAGEHOUR2"))
  if err != nil {
      log.Println("无法解析 WORKMESSAGEHOUR2 环境变量", err)
      workMessageHour2 = 20
  }

  workMessageMinute2, err := strconv.Atoi(os.Getenv("WORKMESSAGEMINUTE2"))
  if err != nil {
      log.Println("无法解析 WORKMESSAGEMINUTE2 环境变量", err)
      workMessageMinute2 = 30
  }

  //標記訊息已發送（才不會一直發送訊息）
  messageSent = true

  return groupID, userNames, groupMemberProfile, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2
}

func remindToWork(event linebot.Event, location time.Location) {

  // 调用初始化群组函数以获取相关参数
  groupID, _, groupMemberProfile, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2 := initializeGroup()

  // 记录第一个Webhook事件，包括群组ID
  log.Printf("Received first Webhook event - Group ID: %s", event.Source.GroupID)

  // 调用其他函数，并传递环境变量的值作为参数
  triggerWorkMessage(bot, groupID, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2, event)
  triggerSumAll(bot, groupID, groupMemberProfile, event)

  // 定时触发 "上班囉" 消息
  go triggerWorkMessage(bot, groupID, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2, event)
} //func remindToWork的結束


// 函数用于计算等待的时间
func calculateWaitTime(targetTime time.Time, location *time.Location) time.Duration {
  now := time.Now().In(location)
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

// 触发 "上班囉" 消息的函数
func triggerWorkMessage(bot *linebot.Client, groupID string, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2 int, event *linebot.Event, location *time.Location) {

  for {
    now := time.Now().In(taipeiLocation)
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
          go triggerSumAll(bot, groupID, groupMemberProfile, event) 
      //(以下內容測試之後要刪除)
      } else if weekday == time.Saturday || weekday == time.Sunday {
      // 如果今天是星期六或星期日，发送 "週六日測試" 訊息
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
      sendMessage(bot, groupID, "週六日測試")

      // 设置定时触发 SumAll 的计时器
      go triggerSumAll(bot, groupID, groupMemberProfile, event) 
      } //以上else if內容測試後要刪除
    }
  }



// 觸發sumall(在發送pushMessage之後的30分鐘)
func triggerSumAll(bot *linebot.Client, groupID string, groupMemberProfile string, event *linebot.Event, location *time.Location) {
  count, err := strconv.Atoi(os.Getenv("SUMALLTRIGGERCOUNT"))
  if err != nil {
    log.Println("無法解析SUMALLTRIGGERCOUNT環境變量", err)
    return
  }

  for i := 0; i < count; i++ {
    time.Sleep(30 * time.Minute)

    // 触发 SumAll
    handleSumAll(event, groupMemberProfile)

    // 更新上次触发 SumAll 的时间
    lastSumAllTriggerTime = time.Now().In(taipeiLocation)
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
        } else if strings.EqualFold(message.Text, ":sum_all") && isGroupEvent(event) {
          handleSumAll(event, groupMemberProfile)
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

func handleSumAll(event *linebot.Event, groupMemberProfile string) {

  // Scroll through all the messages in the chat group (in chronological order).
  oriContext := ""
  q := summaryQueue.ReadGroupInfo(getGroupID(event))
  for _, m := range q {
    // [xxx]: 他講了什麼... 時間
    oriContext = oriContext + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, m.Time.Local().UTC().Format("2006-01-02 15:04:05"))
  }

  // 訊息內先回，再來總結。
  // if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("好的，總結文字已經發給您了"+userName)).Do(); err != nil {
  // 	log.Print(err)
  // }

  // 就是請 ChatGPT 幫你總結
  oriContext = fmt.Sprintf("%s %s", oriContext)
  systemMessage:= fmt.Sprintf("以下你會看到的是一個工作群組中的許多訊息，請將以下内容统整，原則上依照訊息時間排序即可。請用繁體中文回覆，僅需整理內容即可，千萬不要捏造不存在的內容。最後，請幫忙整理出尚未在近5小時內發言的同仁。\n\n目前在群组中的使用者有：%s\n\n%s", groupMemberProfile, oriContext)

  //使用chatgpt.go裡面的 func gptChat 处理 oriContext，同時傳送systemMessage
  reply, err := gptChat(oriContext, systemMessage)
  if err != nil {
    fmt.Printf("ChatGPT error: %v\n", err)
    // 處理錯誤
    return
  }


  // 在群組中使用ReplyToken回覆訊息
  if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前總結如下：\n"+reply+"\n\n Group ID: "+event.Source.GroupID)).Do(); err != nil {
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

func handleStoreMsg(event *linebot.Event, message string, location *time.Location) {
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
    Time:     time.Now().In(taipeiLocation),
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
