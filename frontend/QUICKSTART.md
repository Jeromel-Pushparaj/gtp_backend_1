# Quick Start Guide

Get the Slack Approval Workflow frontend up and running in 3 simple steps!

## Prerequisites

- Node.js 18 or higher
- npm, yarn, or pnpm
- Backend approval service running on `http://localhost:8083`

## 🚀 Quick Start

### Step 1: Install Dependencies

```bash
cd frontend
npm install
```

### Step 2: Start Development Server

```bash
npm run dev
```

The application will automatically open in your browser at `http://localhost:3000`

### Step 3: Test the Application

1. Click on a **Quick Template** button (e.g., "Deployment")
2. Fill in the required fields:
   - **Approver Name**: Enter a name (e.g., "Sarumathi S")
   - **Requester Name**: Enter a name (e.g., "Jeromel Pushparaj")
   - **App Bot User ID**: Enter a bot ID (e.g., "U0AGPDSLH0V")
3. Click **Submit Approval Request**
4. Watch the workflow visualization update in real-time!

## 📝 Example Test Request

Use this data for a quick test:

- **Approver Name**: `Sarumathi S`
- **Requester Name**: `Jeromel Pushparaj`
- **Request Type**: `deployment`
- **Message**: `Please approve production deployment`
- **Request Data**:
  ```json
  {
    "service": "api-gateway",
    "version": "v2.1.0",
    "environment": "production"
  }
  ```
- **Use App DM**: ✓ (checked)
- **App Bot User ID**: `U0AGPDSLH0V`

## 🔧 Configuration

If your backend is running on a different port, create a `.env` file:

```bash
echo "VITE_API_BASE_URL=http://localhost:YOUR_PORT" > .env
```

## 🐛 Troubleshooting

### "Cannot connect to backend"
- Ensure the backend service is running: `http://localhost:8083/health`
- Check the console for CORS errors
- Verify the API URL in `.env`

### "Module not found" errors
```bash
rm -rf node_modules package-lock.json
npm install
```

### Port 3000 already in use
The dev server will automatically use the next available port (3001, 3002, etc.)

## 📚 Next Steps

- Read the full [README.md](./README.md) for detailed documentation
- Explore the [API documentation](../services/approval-service/document/generic_approval.txt)
- Customize the templates in `src/App.tsx`

## 🎯 Key Features to Try

1. **Template System**: Click template buttons to auto-fill forms
2. **JSON Editor**: Add custom request data in JSON format
3. **Debug Mode**: Expand "Request/Response Debug Info" to see raw data
4. **Workflow Tracking**: Watch stages update as the request progresses
5. **Form Validation**: Try submitting with missing required fields

Enjoy! 🎉

