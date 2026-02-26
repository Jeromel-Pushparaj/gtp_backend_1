# Kafka Development Setup

This directory contains Kafka-related infrastructure configuration.

## Quick Start

The Kafka setup is managed via the root `docker-compose.yml` file.

### Start Kafka and Dependencies

```bash
# From the project root
docker-compose up -d
```

### Stop Kafka and Dependencies

```bash
docker-compose down
```

### View Kafka Logs

```bash
docker-compose logs -f kafka
```

## Services Included

- **Zookeeper** (port 2181): Coordination service for Kafka
- **Kafka** (port 9092): Message broker
- **Kafka UI** (port 8090): Web interface for managing Kafka
- **PostgreSQL** (port 5432): Database
- **Redis** (port 6379): Cache

## Kafka Topics

The following topics are used by the platform:

- `jira.trigger.created` - Jira trigger events
- `approval.requested` - Approval request events
- `approval.completed` - Approval completion events
- `service.onboarded` - Service onboarding events
- `chat.request` - Chat request events
- `chat.response` - Chat response events

Topics are auto-created when first used.

## Accessing Kafka UI

Open your browser and navigate to: http://localhost:8090

## Testing Kafka Connection

```bash
# List topics
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092

# Create a test topic
docker exec -it kafka kafka-topics --create --topic test --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1

# Produce a test message
docker exec -it kafka kafka-console-producer --topic test --bootstrap-server localhost:9092

# Consume messages
docker exec -it kafka kafka-console-consumer --topic test --from-beginning --bootstrap-server localhost:9092
```

