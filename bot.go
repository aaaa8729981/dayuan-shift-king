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

// 函数用于计算等待的时间
func calculateWaitTime(targetTime time.Time) time.Duration {
  now := time.Now().In(TaipeiLocation)
  if now.After(targetTime) {
      targetTime = targetTime.Add(24 * time.Hour)
    }
  waitTime := targetTime.Sub(now)
  log.Printf("Now: %s, Target Time: %s, Wait Time: %s\n", now, targetTime, waitTime)
  return targetTime.Sub(now)
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
      sendMessage(bot, groupID, "請各位同仁整理今日完成與未完成的工作項目表，謝謝\n\n上課結束的老師務必掃拖教室\n行政同仁掃拖一二樓公共空間、廁所、垃圾淨空\n教室的環境需要一起維護，請大家完成也在群組與我回報，謝謝", globalEvent)
      log.Printf("發送訊息：請各位同仁整理今日完成與未完成的工作項目表，謝謝\n\n上課結束的老師務必掃拖教室\n行政同仁掃拖一二樓公共空間、廁所、垃圾淨空\n教室的環境需要一起維護，請大家完成也在群組與我回報，謝謝")


      //Check if it's time for the first message
      if now.Hour() >= workMessageHour1 && now.Minute() >= workMessageMinute1 {
        //Set the target time for the next day
        targetTime1 = targetTime1.Add(24 * time.Hour)
      }

      //Check if it's time for the second message
      if now.Hour() >= workMessageHour2 && now.Minute() >= workMessageMinute2 {
        //Set the target time for the next day
        targetTime2 = targetTime2.Add(24 * time.Hour)
      } 
    }
  }
}

// 函数用于发送消息
func sendMessage(bot *linebot.Client, groupID string, message string, globalEvent *linebot.Event) error {
  _, err := bot.PushMessage(groupID, linebot.NewTextMessage(message)).Do()
  if err != nil {
    log.Printf("sendMessage發送消息錯誤：%v\n", err)
    return err
  }

  //在发送消息后触发 triggerSumAll
  //先確認globalEvent的值
  if globalEvent != nil {
    time.AfterFunc(10*time.Minute, func() { 
    log.Printf("10分鐘後觸發 triggerSumAll with groupID: %s, groupMemberProfile: %s, event: %+v\n", groupID, groupMemberProfile, globalEvent)
    triggerSumAll(groupID, groupMemberProfile, globalEvent)
    })
    } else {
    log.Println("globalEvent is nil in triggerSumAll")
  }
return nil
}

// 觸發sumall(在發送pushMessage之後的10分鐘)
func triggerSumAll(groupID string, groupMemberProfile string, globalEvent*linebot.Event) {

  count, err := strconv.Atoi(os.Getenv("SUMALLTRIGGERCOUNT"))
  if err != nil {
    log.Println("無法解析SUMALLTRIGGERCOUNT環境變量", err)
    return
  }

  for i := 0; i < count; i++ {
    //紀錄event的值
    log.Printf("triggerSumAll裡event變量的值： %+v\n", globalEvent)

    log.Printf("等待10分鐘，10分後觸發第 %d 次 SumAll\n", i+1)
    time.Sleep(10 * time.Minute) 

    log.Println("觸發時間：", time.Now().In(TaipeiLocation))
    //確保在 event 變數為 nil 時不執行 handleGroupSumAll 函數，避免了空指針異常。
    if globalEvent == nil {
      log.Println("func triggerSumAll event 變數為nil，不執行handleGroupSumAll")
      return
    }

    // 触发 handleGroupSumAll
    log.Printf("觸發第 %d 次 SumAll，參數：event.ReplyToken=%s, event=%+v, groupMemberProfile=%s\n", i+1, globalEvent.ReplyToken, globalEvent, groupMemberProfile)
    handleGroupSumAll(globalEvent, groupMemberProfile)

    // 更新上次触发 SumAll 的时间
    lastSumAllTriggerTime = time.Now().In(TaipeiLocation)
  }
}

func handleGroupSumAll(globalEvent *linebot.Event, groupMemberProfile string) {
  if len(groupMemberProfile) <= 0 {
    //如果groupMemberProfile為空值，從ENV中獲取GROUPMEMBERPROFILE
    groupMemberProfile = os.Getenv("GROUPMEMBERPROFILE")
    log.Println("handleGroupSumAll從os.Getenv取得GROUPMEMBERPROFILE")
    log.Println(groupMemberProfile)
    log.Println("groupMemberProfile 為空值")
  }
    //添加handleGroupSumAll的log
  log.Printf("handleGroupSumAll: Event Information: %+v\n", globalEvent)
    // Scroll through all the messages in the chat group (in chronological order).

    oriContext := ""
    q := summaryQueue.ReadGroupInfo(getGroupID(globalEvent))
    currentTime := time.Now().In(TaipeiLocation)
    fiveHoursAgo := currentTime.Add(-5 * time.Hour)

    for _, m := range q {
      msgTime := m.Time.In(TaipeiLocation)
      if msgTime.After(fiveHoursAgo) {

      // 只處理過去五小時的訊息
      oriContext = oriContext + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, msgTime.Format("15:04"))
      // [xxx]: 他講了什麼... 時間
    }
    // 就是請 ChatGPT 幫你總結
    oriContext = fmt.Sprintf("%s", oriContext)
    systemMessage:= fmt.Sprintf("以下你會看到的是一個工作群組中的許多訊息，但有部分瑣碎日常資訊讓我無法專心，請幫忙列出重要訊息即可，我關注的是同仁一整天要做什麼事情。然後，請整理出尚未在近1小時內發言的同仁。千萬不要捏造不存在的內容。\n\n目前在群组中的使用者有:%s\n\n", groupMemberProfile)

      //使用chatgpt.go裡面的 func gptChat 处理 oriContext，同時傳送systemMessage
      reply, err := gptChat(oriContext, systemMessage)
      log.Println(oriContext)
      if err != nil {
        fmt.Printf("ChatGPT error: %v\n", err)
    // 處理錯誤
    return
      }
      // 在群組中使用ReplyToken回覆訊息
      if reply == "" {
      reply = "[群組自動統整今日沒有相關訊息]"
    }

  if _, err = bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage("謝謝大家\n\n"+reply)).Do(); err != nil {
    log.Print(err)
    } else {
      //印出reply內容
      log.Printf("handleGroupSumAll回覆的訊息內容: \n%s", reply)
    }
  }
}

func callbackHandler(w http.ResponseWriter, r *http.Request, groupMemberProfile string, workMessageHour1 int, workMessageMinute1 int, workMessageHour2 int, workMessageMinute2 int) {
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
    globalEvent = event // 在 callbackHandler 中更新全域的 globalEvent 變量

    if event.Type == linebot.EventTypeMessage {
      switch message := event.Message.(type) {
      // Handle only on text message
      case *linebot.TextMessage:
        // Directly to ChatGPT
        if strings.Contains(message.Text, ":gpt") {
          // New feature.
          if IsRedemptionEnabled() {
            if stickerRedeemable {
              handleGPT(GPT_Complete, globalEvent, message.Text)
              stickerRedeemable = false
            } else {
              handleRedeemRequestMsg(globalEvent)
            }
          } else {
            // Original one
            handleGPT(GPT_Complete, globalEvent, message.Text)
          }
        } else if strings.Contains(message.Text, ":gpt4") {
          // New feature.
          if IsRedemptionEnabled() {
            if stickerRedeemable {
              handleGPT(GPT_GPT4_Complete, globalEvent, message.Text)
              stickerRedeemable = false
            } else {
              handleRedeemRequestMsg(globalEvent)
            }
          } else {
            // Original one
            handleGPT(GPT_GPT4_Complete, globalEvent, message.Text)
          }
        } else if strings.Contains(message.Text, ":draw") {
          // New feature.
          if IsRedemptionEnabled() {
            if stickerRedeemable {
              handleGPT(GPT_Draw, globalEvent, message.Text)
              stickerRedeemable = false
            } else {
              handleRedeemRequestMsg(globalEvent)
            }
          } else {
            // Original one
            handleGPT(GPT_Draw, globalEvent, message.Text)
          }
        } else if strings.EqualFold(message.Text, ":list_all") && isGroupEvent(event) {
          handleListAll(globalEvent)
        } else if strings.EqualFold(message.Text, ":sum_all") && isGroupEvent(event) {
          handleSumAll(globalEvent, groupMemberProfile)
        } else if isGroupEvent(globalEvent) {
          // 如果聊天機器人在群組中，開始儲存訊息。
          //紀錄groupID的值
          log.Printf("func callbackHandler回傳的message: %+v\n", message.Text)
          log.Printf("func callbackHandler回傳的event: %+v\n", globalEvent.Source.GroupID)
          handleStoreMsg(globalEvent, message.Text)
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
            if _, err = bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage("你的賦能功能啟動了！")).Do(); err != nil {
              log.Print(err)
            }
          }
        }

        if isGroupEvent(globalEvent) {
          // 在群組中，一樣紀錄起來不回覆。
          outStickerResult := fmt.Sprintf("貼圖訊息: %s ", kw)
          handleStoreMsg(globalEvent, outStickerResult)
        } else {
          outStickerResult := fmt.Sprintf("貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerID, message.PackageID, kw, message.Text)

          // 1 on 1 就回覆
          if _, err = bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage(outStickerResult)).Do(); err != nil {
            log.Print(err)
          }
        }
      }
    }
  }
}

func handleSumAll(globalEvent *linebot.Event, groupMemberProfile string) {
  if len(groupMemberProfile) <= 0 {
    //如果groupMemberProfile為空值，從ENV中獲取GROUPMEMBERPROFILE
    groupMemberProfile = os.Getenv("GROUPMEMBERPROFILE")
    log.Println("func handleSumAll: 從os.Getenv取得GROUPMEMBERPROFILE")
    log.Println(groupMemberProfile)
  }

  if len(groupMemberProfile) <= 0 {
    //如果groupMemberProfile仍然為空值，記錄到log
    log.Println("func handleSumAll: groupMemberProfile 為空值")
    return
  }

  // Scroll through all the messages in the chat group (in chronological order).
  oriContext := ""
  q := summaryQueue.ReadGroupInfo(getGroupID(globalEvent))
  currentTime := time.Now().In(TaipeiLocation)
  fiveHoursAgo := currentTime.Add(-5 *time.Hour)

  for _, m := range q {
    msgTime := m.Time.In(TaipeiLocation)
    if msgTime.After(fiveHoursAgo){
      //只處理過去五小時的訊息
      oriContext = oriContext + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, msgTime.Format("15:04"))
      // [xxx]: 他講了什麼... 時間
    }
  }

  // 訊息內先回，再來總結。
  // if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("好的，總結文字已經發給您了"+userName)).Do(); err != nil {
  // 	log.Print(err)
  // }
  if oriContext == "" {
    oriContext = "[統整]今天沒有相關訊息"
  }

  // 就是請 ChatGPT 幫你總結
  oriContext = fmt.Sprintf("%s", oriContext)
  systemMessage:= fmt.Sprintf("以下你會看到的是一個工作群組中的許多訊息，但有部分瑣碎日常資訊讓我無法專心，請幫忙列出重要訊息即可，我關注的是同仁一整天要做什麼事情。然後，請整理出尚未在近1小時內發言的同仁。千萬不要捏造不存在的內容。\n\n目前在群组中的使用者有:%s\n\n", groupMemberProfile)

  //使用chatgpt.go裡面的 func gptChat 处理 oriContext，同時傳送systemMessage
  reply, err := gptChat(oriContext, systemMessage)
  log.Println(oriContext)
  if err != nil {
    fmt.Printf("func handleSumAll: ChatGPT error: %v\n", err)
    // 處理錯誤
    return
  }

  if reply == "" {
    reply = "[統整]今天沒有相關訊息"
  }

  // 在群組中使用ReplyToken回覆訊息
  if _, err = bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage("謝謝大家\n\n"+reply)).Do(); err != nil {
  log.Print(err)
  } else {
    //印出reply內容
    log.Printf("func handleSumAll: 回覆的訊息內容：\n%s", reply)
  }
}

func handleListAll(globalEvent *linebot.Event) {
  reply := ""
  q := summaryQueue.ReadGroupInfo(getGroupID(globalEvent))
  today := time.Now().In(TaipeiLocation)

  for _, m := range q {
    //檢查每條訊息的時間戳是否為今日，並僅統整今日的訊息
    msgTime := m.Time.In(TaipeiLocation)
    if msgTime.Year() == today.Year() && msgTime.YearDay() == today.YearDay() {
      reply = reply + fmt.Sprintf("[%s]: %s . %s\n", m.UserName, m.MsgText, time.Now().In(TaipeiLocation).Format("15:04"))
  }
}

  if reply == "" {
    reply = "[條列]今日沒有相關訊息"
  }

  if _, err := bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
    log.Print(err)
  }
}

func handleGPT(action GPT_ACTIONS, globalEvent *linebot.Event, message string) {
  switch action {
  case GPT_Complete:
    reply := gptGPT3CompleteContext(message)
    if _, err := bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
      log.Print(err)
    }
  case GPT_GPT4_Complete:
    reply := gptGPT4CompleteContext(message)
    if _, err := bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
      log.Print(err)
    }
  case GPT_Draw:
    if reply, err := gptImageCreate(message); err != nil {
      if _, err := bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage("無法正確顯示圖形.")).Do(); err != nil {
        log.Print(err)
      }
    } else {
      if _, err := bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage("根據你的提示，畫出以下圖片："), linebot.NewImageMessage(reply, reply)).Do(); err != nil {
        log.Print(err)
      }
    }
  }

}

func handleRedeemRequestMsg(globalEvent *linebot.Event) {
  // First, obtain the user's Display Name (i.e., the name displayed).
  userName := globalEvent.Source.UserID
  userProfile, err := bot.GetProfile(globalEvent.Source.UserID).Do()
  if err == nil {
    userName = userProfile.DisplayName
  }

  if _, err := bot.ReplyMessage(globalEvent.ReplyToken, linebot.NewTextMessage(userName+":你需要買貼圖，開啟這個功能"), linebot.NewStickerMessage(RedeemStickerPID, RedeemStickerSID)).Do(); err != nil {
    log.Print(err)
  }
}

func handleStoreMsg(globalEvent *linebot.Event, message string) {
  // Get user display name. (It is nick name of the user define.)
  userName := globalEvent.Source.UserID
  userProfile, err := bot.GetProfile(globalEvent.Source.UserID).Do()
  if err == nil {
    userName = userProfile.DisplayName
  }

  // event.Source.GroupID 就是聊天群組的 ID，並且透過聊天群組的 ID 來放入 Map 之中。
  m := MsgDetail{
    MsgText:  message,
    UserName: userName,
    Time:     time.Now().In(TaipeiLocation),
  }
  summaryQueue.AppendGroupInfo(getGroupID(globalEvent), m)
}

func isGroupEvent(globalEvent *linebot.Event) bool {
  return globalEvent.Source.GroupID != "" || globalEvent.Source.RoomID != ""
}

func getGroupID(globalEvent *linebot.Event) string {
  if globalEvent.Source.GroupID != "" {
    return globalEvent.Source.GroupID
  } else if globalEvent.Source.RoomID != "" {
    return globalEvent.Source.RoomID
  }

  return ""
}