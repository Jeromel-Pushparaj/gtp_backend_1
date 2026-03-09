# Slack Approval Workflow Frontend

A modern React application for visualizing and managing Slack approval workflows with real-time status tracking.

## 🚀 Features

- **Interactive Form**: Create approval requests with comprehensive validation
- **Real-time Workflow Visualization**: Track approval progress through all stages
- **Template System**: Quick-start templates for common approval types
- **Debug Mode**: View request/response data for troubleshooting
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **Dark Mode Support**: Automatic dark mode based on system preferences
- **Slack-themed UI**: Clean, modern interface with Slack brand colors

## 📋 Workflow Stages

The application visualizes the complete approval workflow:

1. **Request Created** - Approval request initialized
2. **Published to Kafka** - Message sent to `approval.requested` topic
3. **Sent to Slack DM** - Approval message delivered to approver
4. **Pending Approval** - Waiting for approver action
5. **Approved/Rejected** - Approver has made a decision
6. **Published to Kafka** - Result sent to `approval.completed` topic

## 🛠️ Technology Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Styling
- **Axios** - HTTP client
- **Lucide React** - Icon library

## 📦 Installation

### Prerequisites

- Node.js 18+ and npm/yarn/pnpm
- Backend approval service running on `http://localhost:8083`

### Setup Steps

1. **Install dependencies**:
   ```bash
   npm install
   # or
   yarn install
   # or
   pnpm install
   ```

2. **Configure environment** (optional):
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` to change the API endpoint:
   ```env
   VITE_API_BASE_URL=http://localhost:8083
   ```

3. **Start development server**:
   ```bash
   npm run dev
   # or
   yarn dev
   # or
   pnpm dev
   ```

4. **Open browser**:
   The app will automatically open at `http://localhost:3000`

## 🏗️ Build for Production

```bash
npm run build
# or
yarn build
# or
pnpm build
```

The built files will be in the `dist/` directory.

### Preview Production Build

```bash
npm run preview
# or
yarn preview
# or
pnpm preview
```

## 📝 Usage

### Creating an Approval Request

1. **Fill in the form fields**:
   - **Bot ID** (optional): Slack bot identifier
   - **Approver Name** (required): Name of the person who will approve
   - **Requester Name** (required): Name of the person requesting approval
   - **Request Type** (required): Select from dropdown (deployment, access_request, code_review, other)
   - **Message** (required): Detailed description of what needs approval
   - **Request Data** (optional): Additional JSON data specific to the request type
   - **Use App DM** (checkbox): Send to approver's DM (default: true)
   - **App Bot User ID** (required when Use App DM is checked): Bot user ID for DM

2. **Use Quick Templates** (optional):
   Click on any template button to auto-fill the form with example data:
   - **Deployment**: Production deployment approval
   - **Access Request**: Database access request
   - **Code Review**: Pull request review

3. **Submit the request**:
   Click "Submit Approval Request" button

4. **Monitor progress**:
   Watch the workflow visualization update in real-time

### Request Templates

#### Deployment Approval
```json
{
  "approver_name": "Sarumathi S",
  "requester_name": "Jeromel Pushparaj",
  "request_type": "deployment",
  "message": "Please approve production deployment",
  "request_data": {
    "service": "api-gateway",
    "version": "v2.1.0",
    "environment": "production"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

#### Access Request
```json
{
  "approver_name": "Sarumathi S",
  "requester_name": "Jeromel Pushparaj",
  "request_type": "access_request",
  "message": "Requesting access to production database",
  "request_data": {
    "resource": "production-db",
    "access_level": "read-only",
    "duration": "24 hours"
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

#### Code Review
```json
{
  "approver_name": "Jeromel Pushparaj",
  "requester_name": "Sarumathi S",
  "request_type": "code_review",
  "message": "Please review PR #123 for the new feature",
  "request_data": {
    "pr_number": "123",
    "repository": "backend-api",
    "branch": "feature/new-approval-system",
    "lines_changed": 450
  },
  "use_app_dm": true,
  "app_bot_user_id": "U0AGPDSLH0V"
}
```

## 🎨 UI Components

### ApprovalForm
Main form component for creating approval requests with:
- Form validation
- JSON editor for request data
- Template loading
- Error handling

### WorkflowVisualization
Visual representation of the approval workflow with:
- Stage indicators
- Progress tracking
- Status colors (green=completed, blue=in-progress, yellow=pending, red=rejected)
- Request ID display

### StatusIndicator
Individual stage status component with:
- Icon representation
- Status-based coloring
- Description text

## 🔧 Configuration

### Environment Variables

- `VITE_API_BASE_URL`: Backend API base URL (default: `http://localhost:8083`)

### API Endpoint

The application connects to:
```
POST http://localhost:8083/api/v1/approval/generic
```

## 📁 Project Structure

```
frontend/
├── package.json              # Dependencies and scripts
├── tsconfig.json            # TypeScript configuration
├── vite.config.ts           # Vite configuration
├── tailwind.config.js       # Tailwind CSS configuration
├── postcss.config.js        # PostCSS configuration
├── index.html               # HTML entry point
├── .env                     # Environment variables
├── .env.example             # Environment variables template
├── src/
│   ├── main.tsx            # Application entry point
│   ├── App.tsx             # Main application component
│   ├── index.css           # Global styles
│   ├── components/
│   │   ├── ApprovalForm.tsx           # Form component
│   │   ├── WorkflowVisualization.tsx  # Workflow display
│   │   └── StatusIndicator.tsx        # Status indicator
│   ├── types/
│   │   └── approval.ts     # TypeScript type definitions
│   └── services/
│       └── api.ts          # API client
└── README.md               # This file
```

## 🐛 Troubleshooting

### Backend Connection Issues

If you see connection errors:
1. Ensure the backend service is running on `http://localhost:8083`
2. Check the `.env` file has the correct `VITE_API_BASE_URL`
3. Verify CORS is enabled on the backend

### JSON Validation Errors

If request data JSON is invalid:
1. Use the debug section to view the exact request being sent
2. Validate JSON syntax using a JSON validator
3. Use the template buttons for valid examples

### Build Errors

If you encounter build errors:
1. Delete `node_modules` and reinstall: `rm -rf node_modules && npm install`
2. Clear Vite cache: `rm -rf node_modules/.vite`
3. Ensure you're using Node.js 18+

## 🤝 Contributing

1. Follow the existing code style
2. Use TypeScript for type safety
3. Test all form validations
4. Ensure responsive design works on mobile

## 📄 License

This project is part of the GTP Backend system.

## 🔗 Related Documentation

- [Backend API Documentation](../services/approval-service/README.md)
- [Generic Approval Endpoint](../services/approval-service/document/generic_approval.txt)
- [API Endpoints](../services/approval-service/document/api_end_pints.md)

