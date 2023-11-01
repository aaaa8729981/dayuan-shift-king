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
  var err error
  TaipeiLocation, err = time.LoadLocation("Asia/Taipei")
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

  // 調用 initializeGroup 和 triggerWorkMessage
  initializeGroup()
  go triggerWorkMessage(bot, nil, nil, nil, nil, nil, nil)


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