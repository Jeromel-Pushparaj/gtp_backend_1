package v1

import (
	"encoding/json"
	"log"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
)

type ConsumerService struct {
	kafkaService    *KafkaService
	approvalService *ApprovalService
}

func NewConsumerService(kafkaService *KafkaService, approvalService *ApprovalService) *ConsumerService {
	return &ConsumerService{
		kafkaService:    kafkaService,
		approvalService: approvalService,
	}
}

func (cs *ConsumerService) Start() error {
	topics := []string{constants.KafkaTopicApprovalRequested}
	err := cs.kafkaService.Subscribe(topics)
	if err != nil {
		return err
	}

	log.Printf("Subscribed to Kafka topics: %v", topics)

	go func() {
		for {
			msg, err := cs.kafkaService.Consume()
			if err != nil {
				log.Printf("Error consuming message: %v", err)
				continue
			}
			log.Printf("Received message from topic %s: %s", *msg.TopicPartition.Topic, string(msg.Value))

			var approvalRequest resources.ApprovalRequestMessage
			if err := json.Unmarshal(msg.Value, &approvalRequest); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if err := cs.approvalService.ProcessApprovalRequest(&approvalRequest); err != nil {
				log.Printf("Error processing approval request: %v", err)
				continue
			}
			log.Printf("Successfully processed approval request: %s", approvalRequest.RequestID)
		}
	}()
	return nil
}
