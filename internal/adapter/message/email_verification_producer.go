package message

import (
	"user-service/config"
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
)

type EmailVerificationProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailVerficationProducer(producer sarama.SyncProducer, cfg *config.Config) *EmailVerificationProducer {
	return &EmailVerificationProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    cfg.Topic.NotifEmailVerification,
		},
	}
}
