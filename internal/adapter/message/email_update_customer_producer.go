package message

import (
	"user-service/config"
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
)

type EmailUpdateCustomerProducer struct {
	Producer[*model.UserNotificationEvent]
}

func NewEmailUpdateCustomerProducer(producer sarama.SyncProducer, cfg *config.Config) *EmailUpdateCustomerProducer {
	return &EmailUpdateCustomerProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    cfg.Topic.NotifEmailUpdateCustomer,
		},
	}
}
