package storage

import (
	"context"
	"log"
	"pd-service/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBStorage struct {
	client          *mongo.Client
	db              *mongo.Database
	servicesCol     *mongo.Collection
	metricsCol      *mongo.Collection
}

// NewMongoDBStorage creates a new MongoDB storage instance
func NewMongoDBStorage(mongoURL, dbName string) (*MongoDBStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("✅ Connected to MongoDB successfully")

	db := client.Database(dbName)
	
	return &MongoDBStorage{
		client:      client,
		db:          db,
		servicesCol: db.Collection("services"),
		metricsCol:  db.Collection("metrics"),
	}, nil
}

// Close closes the MongoDB connection
func (s *MongoDBStorage) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}

// StoreService stores a service in MongoDB
func (s *MongoDBStorage) StoreService(service *models.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.servicesCol.InsertOne(ctx, service)
	if err != nil {
		log.Printf("❌ Failed to store service in MongoDB: %v", err)
		return err
	}

	log.Printf("✅ Service stored in MongoDB: %s", service.Name)
	return nil
}

// UpsertService updates or inserts a service in MongoDB based on PDServiceID
func (s *MongoDBStorage) UpsertService(service *models.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PDServiceID as the unique identifier for upsert
	filter := bson.M{"pd_service_id": service.PDServiceID}

	// Update the UpdatedAt timestamp
	service.UpdatedAt = time.Now()

	// If CreatedAt is not set, set it now
	if service.CreatedAt.IsZero() {
		service.CreatedAt = time.Now()
	}

	update := bson.M{
		"$set": service,
		"$setOnInsert": bson.M{
			"created_at": service.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := s.servicesCol.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("❌ Failed to upsert service in MongoDB: %v", err)
		return err
	}

	if result.UpsertedCount > 0 {
		log.Printf("✅ Service inserted in MongoDB: %s", service.Name)
	} else if result.ModifiedCount > 0 {
		log.Printf("✅ Service updated in MongoDB: %s", service.Name)
	} else {
		log.Printf("ℹ️ Service unchanged in MongoDB: %s", service.Name)
	}

	return nil
}

// StoreMetrics stores service metrics in MongoDB
func (s *MongoDBStorage) StoreMetrics(metrics []*models.ServiceMetrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if len(metrics) == 0 {
		return nil
	}

	// Convert to interface slice for InsertMany
	docs := make([]interface{}, len(metrics))
	for i, m := range metrics {
		// Add timestamp to track when metrics were captured
		metricsDoc := map[string]interface{}{
			"service_id":         m.ServiceID,
			"service_name":       m.ServiceName,
			"open_incidents":     m.OpenIncidents,
			"total_incidents":    m.TotalIncidents,
			"high_priority":      m.HighPriority,
			"avg_time_to_resolve": m.AvgTimeToResolve,
			"avg_time_to_respond": m.AvgTimeToRespond,
			"assignee_name":      m.AssigneeName,
			"assignee_slack_id":  m.AssigneeSlackID,
			"last_incident_time": m.LastIncidentTime,
			"captured_at":        time.Now(), // Timestamp when metrics were captured
		}
		docs[i] = metricsDoc
	}

	_, err := s.metricsCol.InsertMany(ctx, docs)
	if err != nil {
		log.Printf("❌ Failed to store metrics in MongoDB: %v", err)
		return err
	}

	log.Printf("✅ Stored %d service metrics in MongoDB", len(metrics))
	return nil
}

// GetServiceByName retrieves a service from MongoDB by name
func (s *MongoDBStorage) GetServiceByName(serviceName string) (*models.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var service models.Service
	filter := bson.M{"name": serviceName}

	err := s.servicesCol.FindOne(ctx, filter).Decode(&service)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Service not found
		}
		log.Printf("❌ Failed to get service from MongoDB: %v", err)
		return nil, err
	}

	return &service, nil
}

