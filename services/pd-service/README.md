# PagerDuty Service Dashboard

A comprehensive web application for managing PagerDuty services with GitHub integration and Slack notifications.

## Features

- 🏠 **Home Panel**: View and manage all your PagerDuty services
- 📊 **Scorecard Panel**: Real-time metrics and analytics with visual charts
- ⚡ **Trigger Panel**: Create test incidents and send Slack notifications
- 🔄 **Multi-Organization Support**: Manage services across different organizations
- 📈 **Visual Analytics**: Charts showing incident trends and metrics

## Prerequisites

- Go 1.21 or higher
- PagerDuty API key
- Slack Bot Token
- GitHub Personal Access Token

## Installation

1. Clone the repository
2. Install Go dependencies:
   ```bash
   go mod download
   ```

## Configuration

The application uses environment variables for configuration. These are already set in the `.env` file:

- `PAGERDUTY_API_KEY`: Your PagerDuty API key
- `SLACK_BOT_TOKEN`: Your Slack bot token
- `GITHUB_PAT`: Your GitHub personal access token
- `DEFAULT_ORG`: Default organization name (teknex-poc)
- `PORT`: Server port (default: 8080)

## Running the Application

1. Start the server:
   ```bash
   go run main.go
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

## Usage

### Adding a Service

1. Click the "Add Service" button on the Home panel
2. Fill in the service details:
   - Service Name: A friendly name for your service
   - PagerDuty Service: Select from your existing PD services
   - GitHub Repository: Select the associated repository
   - On-Call Assignee: Select the Slack user to notify

### Viewing Metrics

1. Navigate to the Scorecard panel
2. View overall metrics and individual service analytics
3. Charts show incident counts and trends
4. Click "Refresh" to update metrics

### Triggering Test Incidents

1. Navigate to the Trigger panel
2. Select a service
3. Enter incident details (title, description, priority)
4. Click "Trigger Incident"
5. The assignee will receive a Slack notification

## API Endpoints

- `GET /api/organizations` - List organizations
- `GET /api/services` - List services
- `POST /api/services` - Create a service
- `GET /api/services/{id}` - Get service details
- `DELETE /api/services/{id}` - Delete a service
- `GET /api/metrics` - Get all service metrics
- `GET /api/services/{id}/metrics` - Get service-specific metrics
- `GET /api/pagerduty/services` - List PagerDuty services
- `GET /api/github/repos` - List GitHub repositories
- `GET /api/slack/users` - List Slack users
- `POST /api/incidents/trigger` - Trigger a test incident

## Architecture

- **Backend**: Go with Gorilla Mux for routing
- **Frontend**: Vanilla JavaScript with Chart.js for visualizations
- **Storage**: In-memory storage with JSON file persistence
- **APIs**: PagerDuty, Slack, and GitHub integrations

## Development

The project structure:
```
pd-service/
├── main.go              # Application entry point
├── config/              # Configuration management
├── models/              # Data models
├── storage/             # Data persistence
├── clients/             # API clients (PD, Slack, GitHub)
├── handlers/            # HTTP handlers
├── frontend/            # Web interface
│   ├── index.html
│   ├── styles.css
│   └── app.js
└── data/                # Data storage
```

## License

MIT

