package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
)

// Kafka представляет интерфейс для работы с Kafka
type Kafka interface {
	SendMessage(message []byte) error
	ConsumeMessages(handler func(message *sarama.ConsumerMessage)) error
	Close() error
}

// KafkaClient реализует Kafka интерфейс
type KafkaClient struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
	topic    string
}

// NewKafkaClient создает новый клиент Kafka
func NewKafkaClient(cfg Config) (Kafka, error) {
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Retry.Max = 5
	producerConfig.Producer.Return.Successes = true

	brokers := make([]string, 0)
	brokers = append(brokers, cfg.Broker)

	producer, err := sarama.NewSyncProducer(brokers, producerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Настройка консьюмера
	consumerConfig := sarama.NewConfig()
	consumer, err := sarama.NewConsumer(brokers, consumerConfig)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	result := &KafkaClient{
		producer: producer,
		consumer: consumer,
		topic:    cfg.Topic,
	}

	return result, nil
}

// SendMessage отправляет сообщение в указанный топик
func (kc *KafkaClient) SendMessage(message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: kc.topic,
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := kc.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Printf("Message sent to topic %s, partition %d at offset %d\n", kc.topic, partition, offset)
	return nil
}

// ConsumeMessages потребляет сообщения из топика
func (kc *KafkaClient) ConsumeMessages(handler func(message *sarama.ConsumerMessage)) error {
	partitionList, err := kc.consumer.Partitions(kc.topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	for _, partition := range partitionList {
		pc, err := kc.consumer.ConsumePartition(kc.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to consume partition: %w", err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for message := range pc.Messages() {
				handler(message)
			}
		}(pc)
	}

	return nil
}

// Close закрывает соединения
func (kc *KafkaClient) Close() error {
	var errs []error

	if err := kc.producer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close producer: %w", err))
	}

	if err := kc.consumer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close consumer: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing kafka client: %v", errs)
	}

	return nil
}
