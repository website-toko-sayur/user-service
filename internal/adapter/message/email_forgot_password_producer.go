package message

import (
	"user-service/internal/core/domain/model"
	"user-service/utils"

	"github.com/IBM/sarama"
)

type EmailForgotPasswordProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailForgotPasswordProducer(producer sarama.SyncProducer) *EmailForgotPasswordProducer {
	return &EmailForgotPasswordProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    utils.EmailForgotPassword,
		},
	}
}
