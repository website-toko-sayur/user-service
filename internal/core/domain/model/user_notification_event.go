package model

import "strconv"

// ini adalah payload event yang akan dikirim ke kafka
type UserNotificationEvent struct {
	ReceiverEmail    string `json:"receiver_email"`
	Message          string `json:"message"`
	ReceiverID       int64  `json:"receiver_id"`
	Subject          string `json:"subject"`
	NotificationType string `json:"notification_type"`
}

// bagian ini dipakai sebagai kafka key.
// karena struct disini memiliki method GetId(), barulah *UserNotificationEvent memenuhi interface: model.Event
func (u *UserNotificationEvent) GetId() string {
	return strconv.Itoa(int(u.ReceiverID))
}
