package main

import "time"

type GroupDB interface {
	ReadGroupInfo(string) GroupData
	AppendGroupInfo(string, MsgDetail)
}
type MsgDetail struct {
	MsgText  string
	UserName string
	Time     time.Time
  QuotedMessageID string  //新增處理使用者回覆訊息的邏輯
}

type GroupData []MsgDetail
