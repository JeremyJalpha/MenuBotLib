package menubotlib

import (
	"database/sql"
	"time"
)

type Pricelist struct {
	PrlstPreamble string
	Catalogue     []CatalogueSelection
}

type ConversationContext struct {
	UserInfo     UserInfo
	UserExisted  bool
	Pricelist    Pricelist
	CurrentOrder CustomerOrder
	MessageBody  string
	DBReadTime   time.Time
}

func NewConversationContext(db *sql.DB, senderNumber, messagebody string, prlst Pricelist, isAutoInc bool) *ConversationContext {
	userInfo, curOrder, userExisted := NewUserInfo(db, senderNumber, isAutoInc)
	userInfo.CellNumber = senderNumber
	context := &ConversationContext{
		UserInfo:     userInfo,
		UserExisted:  userExisted,
		Pricelist:    prlst,
		CurrentOrder: curOrder,
		MessageBody:  messagebody,
		DBReadTime:   time.Now(),
	}

	return context
}
