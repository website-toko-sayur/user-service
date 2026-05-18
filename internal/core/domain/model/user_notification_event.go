package model

import "strconv"

type UserNotificationEvent struct {
	UserID    int64  `json:"user_id"`
	Recipient string `json:"recipient"` // isi email
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

type NotificationEmailverification struct {
}

func (u *UserNotificationEvent) GetId() string {
	return strconv.Itoa(int(u.UserID))
}
