package message

import (
	"encoding/json"
	"user-service/internal/core/domain/model"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

// ini adalah generic producer
// T adalah tipe data event
// T harus memenuhi interface model.Event, jadi tidak semua tipe data boleh digunakan
// kenapa dibuat generic, supaya tidak perlu membuat producer kafka dari nol untuk setiap jenis event
// dengan generic logic kafka-nya cukup dibuat sekali
type Producer[T model.Event] struct {
	Producer sarama.SyncProducer
	Topic    string
}

func (p *Producer[T]) GetTopic() *string {
	return &p.Topic
}

func (p *Producer[T]) Send(event T) error {
	// event diubah menjadi json
	value, err := json.Marshal(event)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.message.Producer.Send").
			Msg("Failed to marshal event")
		return err
	}

	message := &sarama.ProducerMessage{
		Topic: p.Topic,
		// menggunakan key untuk menentukan partition
		// dengan adanya key, maka message dengan key yang sama akan masuk ke partition yang sama,
		// sehingga ordering untuk key tersebut dapat dipertahankan.
		// [karena satu consumer akan membaca satu partition]
		// [barulah kalau consumer itu dihentikan maka akan dikirim ke consumer lain di consumer-group yang sama]
		Key:   sarama.StringEncoder(event.GetId()),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.Producer.SendMessage(message)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "internal.adapter.message.Producer.Send").
			Msg("Failed to produce message")
		return err
	}

	log.Debug().
		Str("source", "internal.adapter.message.Producer.Send").
		Str("topic", p.Topic).
		Int32("partition", partition).
		Int64("offset", offset).
		Msg("Kafka message published")

	return nil
}
