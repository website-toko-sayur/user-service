package message

import (
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
)

type UserNotificationProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewUserNotificationProducer(producer sarama.SyncProducer, topic string) *UserNotificationProducer {
	return &UserNotificationProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    topic,
		},
	}
}
