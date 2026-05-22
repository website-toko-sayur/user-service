package message

import (
	"user-service/config"
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
)

type EmailForgotPasswordProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailForgotPasswordProducer(producer sarama.SyncProducer, cfg *config.Config) *EmailForgotPasswordProducer {
	return &EmailForgotPasswordProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    cfg.Topic.NotifEmailForgotPassword,
		},
	}
}
