package v1

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaService struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
	brokers  string
}

func NewKafkaService(brokers, groupID string) (*KafkaService, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":     brokers,
		"broker.address.family": "v4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     brokers,
		"group.id":              groupID,
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
	})
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	log.Printf("Kafka service initialized with brokers: %s", brokers)
	return &KafkaService{
		producer: producer,
		consumer: consumer,
		brokers:  brokers,
	}, nil
}

func (k *KafkaService) Publish(topic string, key string, message interface{}) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          messageBytes,
	}, deliveryChan)

	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %w", m.TopicPartition.Error)
	}

	log.Printf("Message delivered to topic %s [partition %d] at offset %v",
		*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)

	return nil
}

func (k *KafkaService) Subscribe(topics []string) error {
	return k.consumer.SubscribeTopics(topics, nil)
}

func (k *KafkaService) Consume() (*kafka.Message, error) {
	return k.consumer.ReadMessage(-1)
}

func (k *KafkaService) Close() {
	if k.producer != nil {
		k.producer.Close()
	}
	if k.consumer != nil {
		k.consumer.Close()
	}
	log.Println("Kafka service closed")
}
