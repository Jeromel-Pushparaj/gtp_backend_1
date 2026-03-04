package v1

import (
	"encoding/json"
	"log"

	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
)

type BusinessConsumerService struct {
	consumerKafka *KafkaService
	producerKafka *KafkaService
}

func NewBusinessConsumerService(brokers string, groupID string, producerKafka *KafkaService) (*BusinessConsumerService, error) {
	consumerKafka, err := NewKafkaService(brokers, groupID)
	if err != nil {
		return nil, err
	}

	return &BusinessConsumerService{
		consumerKafka: consumerKafka,
		producerKafka: producerKafka,
	}, nil
}

func (bcs *BusinessConsumerService) Start() error {
	topics := []string{constants.KafkaTopicApprovalCompleted}
	err := bcs.consumerKafka.Subscribe(topics)
	if err != nil {
		return err
	}

	log.Printf("Business Consumer subscribed to Kafka topics: %v", topics)

	go func() {
		for {
			msg, err := bcs.consumerKafka.Consume()
			if err != nil {
				log.Printf("Business Consumer: Error consuming message: %v", err)
				continue
			}
			log.Printf("Business Consumer: Received message from topic %s: %s", *msg.TopicPartition.Topic, string(msg.Value))

			var completedMsg resources.ApprovalCompletedMessage
			if err := json.Unmarshal(msg.Value, &completedMsg); err != nil {
				log.Printf("Business Consumer: Error unmarshaling message: %v", err)
				continue
			}

			if err := bcs.processApprovalCompletion(&completedMsg); err != nil {
				log.Printf("Business Consumer: Error processing approval completion: %v", err)
				continue
			}
			log.Printf("Business Consumer: Successfully processed approval completion: %s", completedMsg.RequestID)
		}
	}()
	return nil
}

func (bcs *BusinessConsumerService) Close() {
	if bcs.consumerKafka != nil {
		bcs.consumerKafka.Close()
	}
}

func (bcs *BusinessConsumerService) processApprovalCompletion(msg *resources.ApprovalCompletedMessage) error {
	log.Printf("Business Consumer: Processing approval completion for request: %s, approved: %t", msg.RequestID, msg.Approved)

	var actionTopic string
	var actionMessage resources.ActionMessage

	if msg.Approved {
		actionTopic = constants.KafkaTopicActionExecuted
		actionMessage = resources.ActionMessage{
			RequestID:   msg.RequestID,
			Action:      "execute",
			Status:      "executed",
			ProcessedBy: msg.ProcessedBy,
			ProcessedAt: msg.ProcessedAt,
			RequestData: msg.RequestData,
			Message:     "Action executed successfully based on approval",
		}
		log.Printf("Business Consumer: Executing business action for request: %s", msg.RequestID)
	} else {
		actionTopic = constants.KafkaTopicActionRejected
		actionMessage = resources.ActionMessage{
			RequestID:   msg.RequestID,
			Action:      "reject",
			Status:      "rejected",
			ProcessedBy: msg.ProcessedBy,
			ProcessedAt: msg.ProcessedAt,
			RequestData: msg.RequestData,
			Message:     "Action rejected based on approval decision",
			Reason:      msg.Reason,
		}
		log.Printf("Business Consumer: Rejecting business action for request: %s, reason: %s", msg.RequestID, msg.Reason)
	}

	err := bcs.producerKafka.Publish(actionTopic, msg.RequestID, actionMessage)
	if err != nil {
		return err
	}

	log.Printf("Business Consumer: Published action message to topic: %s for request: %s", actionTopic, msg.RequestID)
	return nil
}
