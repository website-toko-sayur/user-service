package message

import (
	"user-service/internal/core/domain/model"
	"user-service/utils"

	"github.com/IBM/sarama"
)

type EmailUpdateCustomerProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailUpdateCustomerProducer(producer sarama.SyncProducer) *EmailUpdateCustomerProducer {
	return &EmailUpdateCustomerProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    utils.NOTIF_EMAIL_UPDATE_CUSTOMER,
		},
	}
}
