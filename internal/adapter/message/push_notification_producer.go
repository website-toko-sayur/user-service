package message

import (
	"user-service/internal/core/domain/model"
	"user-service/utils"

	"github.com/IBM/sarama"
)

type PushNotificationProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewPushNotificationProducer(producer sarama.SyncProducer) *PushNotificationProducer {
	return &PushNotificationProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    utils.PUSH_NOTIF,
		},
	}
}
