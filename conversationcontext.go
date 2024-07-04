package menubotlib

import (
	"database/sql"
	"time"
)

type ConversationContext struct {
	UserInfo     UserInfo
	UserExisted  bool
	CurrentOrder CustomerOrder
	MessageBody  string
	DBReadTime   time.Time
}

func NewConversationContext(db *sql.DB, senderNumber, messagebody string, isAutoInc bool) *ConversationContext {
	userInfo, curOrder, userExisted := NewUserInfo(db, senderNumber, isAutoInc)
	context := &ConversationContext{
		UserInfo:     userInfo,
		UserExisted:  userExisted,
		CurrentOrder: curOrder,
		MessageBody:  messagebody,
		DBReadTime:   time.Now(),
	}

	return context
}
