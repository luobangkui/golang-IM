package message

type Msg struct {
	MsgId int64 `json:"msg_id"`

	Content string	`json:"content"`

	SenderId int64 `json:"sender_id"`

	RecipientId int64	`json:"recipient_id"`

	// 1:login 2:chat
	MsgType int `json:"msg_type"`

	CreateTime int `json:"create_time"`

}
