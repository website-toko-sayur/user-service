package message

import (
	"user-service/config"
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
)

type PushNotificationProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewPushNotificationProducer(producer sarama.SyncProducer, cfg *config.Config) *PushNotificationProducer {
	return &PushNotificationProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    cfg.Topic.PushNotif,
		},
	}
}
