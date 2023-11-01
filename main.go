// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "time"
  "github.com/line/line-bot-sdk-go/v7/linebot"
  "github.com/sashabaranov/go-openai"
  "strconv"
)

var bot *linebot.Client
var client *openai.Client
var summaryQueue GroupDB
var stickerRedeemable bool
var enableRedeem string
var TaipeiLocation *time.Location  //TaipeiLocation redeclared in this block

const RedeemStickerPID = "789"
const RedeemStickerSID = "10856"

type GPT_ACTIONS int

const (
  GPT_Complete      GPT_ACTIONS = 0
  GPT_Draw          GPT_ACTIONS = 1
  GPT_Whister       GPT_ACTIONS = 2
  GPT_GPT4_Complete GPT_ACTIONS = 3
)

var messageSent bool

func main() {
// 先初始化 TaipeiLocation
TaipeiLocation, err := time.LoadLocation("Asia/Taipei")
if err != nil {
    log.Fatal("无法设置时区：", err)
  }
  
  stickerRedeemable = false

  // Enable new feature (YES, default no)
  enableRedeem = os.Getenv("REDEEM_ENABLE")

  //  If DABTASE_URL is preset, create PostGresSQL; otherwise, create Mem DB.
  pSQL := os.Getenv("DATABASE_URL")
  if pSQL != "" {
    summaryQueue = NewPGSql(pSQL)
  } else {
    summaryQueue = NewMemDB()
  }

  bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
  log.Println("Bot:", bot, " err:", err)

  // 获取工作小时和分钟值
  workMessageHour1, err := strconv.Atoi(os.Getenv("WORKMESSAGEHOUR1"))
  if err != nil {
    log.Println("无法解析 WORKMESSAGEHOUR1 环境变量", err)
    workMessageHour1 = 11 // 设置默认值
  }

  workMessageMinute1, err := strconv.Atoi(os.Getenv("WORKMESSAGEMINUTE1"))
  if err != nil {
    log.Println("无法解析 WORKMESSAGEMINUTE1 环境变量", err)
    workMessageMinute1 = 0 // 设置默认值
  }

  workMessageHour2, err := strconv.Atoi(os.Getenv("WORKMESSAGEHOUR2"))
  if err != nil {
    log.Println("无法解析 WORKMESSAGEHOUR2 环境变量", err)
    workMessageHour2 = 20 // 设置默认值
  }

  workMessageMinute2, err := strconv.Atoi(os.Getenv("WORKMESSAGEMINUTE2"))
  if err != nil {
    log.Println("无法解析 WORKMESSAGEMINUTE2 环境变量", err)
    workMessageMinute2 = 30 // 设置默认值
  }

  // 获取其他必要的变量，例如 groupID
  groupID := os.Getenv("LINEBOTGROUP_ID")

  // 调用 initializeGroup 和 triggerWorkMessage，传递 TaipeiLocation
  groupID, userNames, groupMemberProfile, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2 := initializeGroup(TaipeiLocation)
  go triggerWorkMessage(bot, groupID, workMessageHour1, workMessageMinute1, workMessageHour2, workMessageMinute2, nil)


  port := os.Getenv("PORT")
  apiKey := os.Getenv("ChatGptToken")

  if apiKey != "" {
    client = openai.NewClient(apiKey)
  }

  http.HandleFunc("/callback", callbackHandler)
  addr := fmt.Sprintf(":%s", port)
  http.ListenAndServe(addr, nil)
}

func IsRedemptionEnabled() bool {
  return enableRedeem == "YES"
}
