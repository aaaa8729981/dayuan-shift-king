package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/line/line-bot-sdk-go/v7/linebot"
)

// 定義一個全局變量用於記錄上次觸發sumall的時間
var lastSumAllTriggerTime time.Time
var groupMemberProfile string // 將 groupMemberProfile 變數宣告為全域變數

func initializeGroup() (string, string){ //func main調用之後取得groupID, groupMemberProfile
  if messageSent {
    // 消息已经发送，不需要再发送
    return "", ""
  }

  // 从 env 中获取 LINEBOTGROUP_ID
  groupIDFromEnv := os.Getenv("LINEBOTGROUP_ID")

  var groupID string
  if groupIDFromEnv != "" {
    groupID = groupIDFromEnv //groupID是從env裡面設定的
    log.Println("groupID:", groupID)  
  } else {
    // 如果ENV檔案中groupID為空值，寫入log並跳過這功能
    log.Println("未设置 Env 的 groupID")
    return "", ""
  }

  // 透过 groupID 取得指定群组成员列表 (userID)
  memberIDsResponse, err := bot.GetGroupMemberIDs(groupID, "").Do()
  var userNames []string

  if err != nil {
      log.Println("透過env取得群組成員列表:") //取得失敗時，不判定為錯誤。userNames繼續維持為空值
  } else {
    for _, userID := range memberIDsResponse.MemberIDs {
      userNames = append(userNames, userID)
    }
    log.Println("群組成員 id 列表:", userNames)

    // 僅在有成功取得成員列表時使用 userNames 變數
    groupMemberProfile = ""
    if len(userNames) > 0 { //如果userNames長度>0表示有成功
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
  log.Println("群組成員名稱:", groupMemberProfile)

    }
  //標記訊息已發送（才不會一直發送訊息）
  messageSent = true

  return groupID, groupMemberProfile
}

// 函数用于计算等待的时间
func calculateWaitTime(targetTime time.Time) time.Duration {
  now := time.Now().In(TaipeiLocation)
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

// 發送 "上班囉" 消息的函数（在func main一開始就調用此函數）
func triggerWorkMessage(bot *linebot.Client, groupID string, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2 int, event *linebot.Event, groupMemberProfile string) {
  fmt.Printf("groupID: %s, workMessageHour1: %d, workMessageMinute1: %d, workMessageHour2: %d, workMessageMinute2: %d, event: %+v, groupMemberProfile: %s\n", groupID, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2, event, groupMemberProfile)

  for {
    now := time.Now().In(TaipeiLocation)
    weekday := now.Weekday()

    // 僅在星期一到星期五执行
    if weekday >= time.Monday && weekday <= time.Friday {
      targetTime1 := time.Date(now.Year(), now.Month(), now.Day(), workMessageHour1, workMessageMinute1, 0, 0, TaipeiLocation)
          targetTime2 := time.Date(now.Year(), now.Month(), now.Day(), workMessageHour2, workMessageMinute2, 0, 0, TaipeiLocation)

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
      log.Println("發送訊息：請各位同仁整理今日工作項目表，謝謝")
      sendMessage(bot, groupID, "請各位同仁整理今日工作項目表，謝謝")
      // 在觸發 triggerSumAll 前添加日誌
      log.Println("等待時間已過，觸發 triggerSumAll")
      // 使用 time.AfterFunc 安排在30分鐘後觸發 triggerSumAll 函數
      time.AfterFunc(3*time.Minute, func() { //先調整為3分鐘測試用
        triggerSumAll(groupID, groupMemberProfile, event)
      })
      return // 退出當前循環，等待下一輪檢查
      } 
    }
  }

// 觸發sumall(在發送pushMessage之後的30分鐘)
func triggerSumAll(groupID string, groupMemberProfile string, event*linebot.Event) {
  count, err := strconv.Atoi(os.Getenv("SUMALLTRIGGERCOUNT"))
  if err != nil {
    log.Println("無法解析SUMALLTRIGGERCOUNT環境變量", err)
    return
  }

  for i := 0; i < count; i++ {
    log.Printf("等待30分鐘，然後觸發第 %d 次 SumAll\n", i+1)
    time.Sleep(1 * time.Minute) //先調整為3分鐘測試用

    //確保在 event 變數為 nil 時不執行 handleGroupSumAll 函數，避免了空指針異常。
    if event == nil {
      log.Println("event 變數為nil")
      return
    }

    //紀錄event的值
    log.Printf("triggerSumAll裡event變量的值： %+v\n", event)

    // 触发 handleGroupSumAll
    log.Printf("觸發第 %d 次 SumAll\n", i+1)
    handleGroupSumAll(event.ReplyToken, event, groupMemberProfile)

    // 更新上次触发 SumAll 的时间
    lastSumAllTriggerTime = time.Now().In(TaipeiLocation)
  }
}

func handleGroupSumAll(replyToken string, event *linebot.Event, groupMemberProfile string) {
  if len(groupMemberProfile) <= 0 {
    //如果groupMemberProfile為空值，從ENV中獲取GROUPMEMBERPROFILE
    groupMemberProfile = os.Getenv("GROUPMEMBERPROFILE")
    log.Println("handleGroupSumAll從os.Getenv取得GROUPMEMBERPROFILE")
    log.Println(groupMemberProfile)
    log.Println("groupMemberProfile 為空值")
  }
    // Scroll through all the messages in the chat group (in chronological order).
    oriContext := ""
    q := summaryQueue.ReadGroupInfo(getGroupID(event))
    for _, m := range q {
      // [xxx]: 他講了什麼... 時間
      oriContext = oriContext + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, time.Now().In(TaipeiLocation).Format("2006-01-02 15:04:05"))
    }
    // 就是請 ChatGPT 幫你總結
    oriContext = fmt.Sprintf("%s", oriContext)
    systemMessage:= fmt.Sprintf("以下你會看到的是一個工作群組中的許多訊息，請幫忙整理出尚未在近1小時內發言的同仁。千萬不要捏造不存在的內容。\n\n目前在群组中的使用者有：%s\n\n", groupMemberProfile)

      //使用chatgpt.go裡面的 func gptChat 处理 oriContext，同時傳送systemMessage
      reply, err := gptChat(oriContext, systemMessage)
      log.Println(oriContext)
      if err != nil {
        fmt.Printf("ChatGPT error: %v\n", err)
    // 處理錯誤
    return
      }
        // 在群組中使用ReplyToken回覆訊息
  if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("謝謝大家\n"+reply+"\n\nGroup ID: "+event.Source.GroupID)).Do(); err != nil {
    log.Print(err)
    } else {
      //印出reply內容
      log.Printf("handleGroupSumAll回覆的訊息內容：\n%s", reply)
}
}

func callbackHandler(w http.ResponseWriter, r *http.Request, groupMemberProfile string) {
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
          triggerWorkMessage(bot, event.Source.GroupID, 11, 0, 20, 30, event, "")
          triggerSumAll(event.Source.GroupID, groupMemberProfile, event)
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
  if len(groupMemberProfile) <= 0 {
    //如果groupMemberProfile為空值，從ENV中獲取GROUPMEMBERPROFILE
    groupMemberProfile = os.Getenv("GROUPMEMBERPROFILE")
    log.Println("從os.Getenv取得GROUPMEMBERPROFILE")
    log.Println(groupMemberProfile)
  }

  if len(groupMemberProfile) <= 0 {
    //如果groupMemberProfile仍然為空值，記錄到log
    log.Println("groupMemberProfile 為空值")
    return
  }

  // Scroll through all the messages in the chat group (in chronological order).
  oriContext := ""
  q := summaryQueue.ReadGroupInfo(getGroupID(event))
  for _, m := range q {
    // [xxx]: 他講了什麼... 時間
    oriContext = oriContext + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, time.Now().In(TaipeiLocation).Format("2006-01-02 15:04:05"))
  }

  // 訊息內先回，再來總結。
  // if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("好的，總結文字已經發給您了"+userName)).Do(); err != nil {
  // 	log.Print(err)
  // }

  // 就是請 ChatGPT 幫你總結
  oriContext = fmt.Sprintf("%s", oriContext)
  systemMessage:= fmt.Sprintf("以下你會看到的是一個工作群組中的許多訊息，請幫忙整理出尚未在近1小時內發言的同仁。千萬不要捏造不存在的內容。\n\n目前在群组中的使用者有：%s\n\n", groupMemberProfile)

  //使用chatgpt.go裡面的 func gptChat 处理 oriContext，同時傳送systemMessage
  reply, err := gptChat(oriContext, systemMessage)
  log.Println(oriContext)
  if err != nil {
    fmt.Printf("ChatGPT error: %v\n", err)
    // 處理錯誤
    return
  }

  // 在群組中使用ReplyToken回覆訊息
  if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("謝謝大家\n"+reply+"\n\nGroup ID: "+event.Source.GroupID)).Do(); err != nil {
  log.Print(err)
  } else {
    //印出reply內容
    log.Printf("回覆的訊息內容：\n%s", reply)
  }
}

func handleListAll(event *linebot.Event) {
  reply := ""
  q := summaryQueue.ReadGroupInfo(getGroupID(event))
  for _, m := range q {
    reply = reply + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, time.Now().In(TaipeiLocation).Format("2006-01-02 15:04:05"))
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
    Time:     time.Now().In(TaipeiLocation),
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