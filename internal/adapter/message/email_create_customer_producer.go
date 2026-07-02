package message

import (
	"user-service/config"
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
)

type EmailCreateCustomerProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailCreateCustomerProducer(producer sarama.SyncProducer, cfg *config.Config) *EmailCreateCustomerProducer {
	return &EmailCreateCustomerProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    cfg.Topic.NotifEmailCreateCustomer,
		},
	}
}
