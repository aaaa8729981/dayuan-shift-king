package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

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
				} else if strings.EqualFold(message.Text, "648599") && isGroupEvent(event) {
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
	userProfile, err := bot.GetProfile(event.Source.UserID).Do()
	if err == nil {
		userName = userProfile.DisplayName
	}

	// 訊息內先回，再來總結。
	// if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("好的，總結文字已經發給您了"+userName)).Do(); err != nil {
	// 	log.Print(err)
	// }

	// 就是請 ChatGPT 幫你總結
	oriContext = fmt.Sprintf("下面的許多訊息是一個排班工作的交換工作時間群組，內容會包含想換班的時間日期、上班時間等資訊，雖然包含許多特定的名詞，但沒關係。請嘗試依照這種範例方式整理資料:10/23（一）範例先生想要换早班\n範例小姐13B想換晚班\n\n10/24（二）\n範例先生 15A想要換晚班。（範例請勿加到回覆內容中）（內容一定會包含日期，請協助格式整理。一個訊息中常包含多個日期，請將日期分開）如果你看不懂資料，請列在最後面，不要嘗試修改或捏造。資料請依照日期先後排序 `%s`", oriContext)
	reply := gptGPT3CompleteContext(oriContext)

	// 因為 ChatGPT 可能會很慢，所以這邊後來用 SendMsg 來發送私訊給使用者。
	if _, err = bot.PushMessage(event.Source.UserID, linebot.NewTextMessage(reply)).Do(); err != nil {
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
