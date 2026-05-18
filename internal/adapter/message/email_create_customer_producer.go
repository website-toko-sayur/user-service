package message

import (
	"user-service/internal/core/domain/model"
	"user-service/utils"

	"github.com/IBM/sarama"
)

type EmailCreateCustomerProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailCreateCustomerProducer(producer sarama.SyncProducer) *EmailCreateCustomerProducer {
	return &EmailCreateCustomerProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    utils.EmailCreateCustomer,
		},
	}
}
