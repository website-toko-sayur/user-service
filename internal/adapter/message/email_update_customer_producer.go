package message

import (
	"user-service/config"
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
)

// ini berarti:
// EmailUpdateCustomerProducer memiliki sebuah generic Producer yang khusus menerima *model.UserNotificationEvent
// jadi secara konsep:
/**
EmailUpdateCustomerProducer
           │
           └── Producer[*UserNotificationEvent]
                        │
                        ├── sarama.SyncProducer
                        └── Topic
*/
type EmailUpdateCustomerProducer struct {
	Producer[*model.UserNotificationEvent]
}

// ini membuat Producer khusus untuk email update customer
func NewEmailUpdateCustomerProducer(producer sarama.SyncProducer, cfg *config.Config) *EmailUpdateCustomerProducer {
	return &EmailUpdateCustomerProducer{
		Producer: Producer[*model.UserNotificationEvent]{
			Producer: producer,
			Topic:    cfg.Topic.NotifEmailUpdateCustomer,
		},
	}
}
