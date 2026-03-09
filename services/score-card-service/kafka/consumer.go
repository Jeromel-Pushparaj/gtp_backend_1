package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	consumer sarama.ConsumerGroup
	topic    string
	handler  MessageHandler
}

// MessageHandler is a function that processes consumed messages
type MessageHandler func(message []byte) error

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, groupID, topic string, handler MessageHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Kafka consumer connected to %v", brokers)

	return &Consumer{
		consumer: consumer,
		topic:    topic,
		handler:  handler,
	}, nil
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	consumerHandler := &consumerGroupHandler{
		handler: c.handler,
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Stopping Kafka consumer...")
			return nil
		default:
			if err := c.consumer.Consume(ctx, []string{c.topic}, consumerHandler); err != nil {
				log.Printf("❌ Error consuming messages: %v", err)
				return err
			}
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler MessageHandler
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("📬 Received message from topic %s, partition %d, offset %d",
			message.Topic, message.Partition, message.Offset)

		if err := h.handler(message.Value); err != nil {
			log.Printf("❌ Error processing message: %v", err)
			// Continue processing other messages even if one fails
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}

// ScoreRequestEvent represents a score calculation request event
type ScoreRequestEvent struct {
	ServiceName string                 `json:"service_name"`
	Metrics     map[string]interface{} `json:"metrics"`
}

// ParseScoreRequestEvent parses a score request event from JSON
func ParseScoreRequestEvent(data []byte) (*ScoreRequestEvent, error) {
	var event ScoreRequestEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

