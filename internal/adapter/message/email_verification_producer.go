package message

import (
	"user-service/internal/core/domain/model"
	"user-service/utils"

	"github.com/IBM/sarama"
)

type EmailVerificationProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailVerficationProducer(producer sarama.SyncProducer) *EmailVerificationProducer {
	return &EmailVerificationProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    utils.NOTIF_EMAIL_VERIFICATION,
		},
	}
}
