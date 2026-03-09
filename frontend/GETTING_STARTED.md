# Getting Started with Slack Approval Workflow Frontend

Welcome! This guide will get you up and running in under 5 minutes.

## 🎯 What You're Building

A beautiful, modern web interface to create and track Slack approval requests through a complete Kafka-driven workflow.

## 📋 What You Need

1. **Node.js 18+** - [Download here](https://nodejs.org/)
2. **Backend Service** - Running on `http://localhost:8083`
3. **5 minutes** - That's it!

## 🚀 Three Steps to Success

### Step 1: Install (2 minutes)

Open your terminal in the `frontend/` directory and run:

```bash
./setup.sh
```

Or manually:

```bash
npm install
```

**What this does:**
- Installs React, TypeScript, Vite, Tailwind CSS, and other dependencies
- Creates `.env` file with default configuration
- Sets up the development environment

### Step 2: Start (30 seconds)

```bash
npm run dev
```

**What happens:**
- Development server starts on `http://localhost:3000`
- Browser opens automatically
- You see the Slack Approval Workflow interface

### Step 3: Test (2 minutes)

1. **Click the "Deployment" template button** - Form auto-fills with example data

2. **Fill in the required fields:**
   - Approver Name: `Sarumathi S`
   - Requester Name: `Jeromel Pushparaj`
   - App Bot User ID: `U0AGPDSLH0V`

3. **Click "Submit Approval Request"**

4. **Watch the magic happen:**
   - ✅ Request Created
   - ✅ Published to Kafka
   - ✅ Sent to Slack DM
   - 🔵 Pending Approval (waiting for approver)

**Success!** You've just created your first approval request! 🎉

## 🎨 What You'll See

### Left Side: Approval Form
- Quick template buttons for common requests
- Form fields with validation
- JSON editor for custom data
- Submit button with loading state

### Right Side: Workflow Visualization
- Real-time progress tracking
- 6 workflow stages with status indicators
- Color-coded states (green=done, blue=active, yellow=pending)
- Request ID display

### Bottom: Debug Section
- View raw request/response data
- Helpful for troubleshooting
- Collapsible to save space

## 📝 Try These Examples

### Example 1: Deployment Approval
```
Template: Click "Deployment"
Approver: Your manager's name
Requester: Your name
Bot ID: U0AGPDSLH0V
```

### Example 2: Access Request
```
Template: Click "Access Request"
Approver: Database admin's name
Requester: Your name
Bot ID: U0AGPDSLH0V
```

### Example 3: Code Review
```
Template: Click "Code Review"
Approver: Senior developer's name
Requester: Your name
Bot ID: U0AGPDSLH0V
```

## 🔧 Configuration

### Change API Endpoint

Edit `.env` file:
```env
VITE_API_BASE_URL=http://your-backend-url:port
```

Restart the dev server after changing.

### Customize Templates

Edit `src/App.tsx` and modify the `requestTemplates` array.

## 🐛 Troubleshooting

### "Cannot connect to backend"

**Check 1:** Is backend running?
```bash
curl http://localhost:8083/health
```

**Check 2:** Is the URL correct in `.env`?
```bash
cat .env
```

**Check 3:** CORS enabled on backend?
- Backend must allow requests from `http://localhost:3000`

### "Module not found" errors

```bash
rm -rf node_modules package-lock.json
npm install
```

### Port 3000 already in use

Vite will automatically use the next available port (3001, 3002, etc.)

### TypeScript errors

```bash
npm run build
```

If errors persist, check `tsconfig.json` is correct.

## 📚 Learn More

- **Full Documentation**: [README.md](./README.md)
- **Example Requests**: [EXAMPLES.md](./EXAMPLES.md)
- **Architecture**: [ARCHITECTURE.md](./ARCHITECTURE.md)
- **Verification**: [CHECKLIST.md](./CHECKLIST.md)

## 🎯 Next Steps

1. ✅ **You're already running!** - Keep exploring the interface
2. 📖 **Read the examples** - See different request types in [EXAMPLES.md](./EXAMPLES.md)
3. 🎨 **Customize** - Modify templates and styling to match your needs
4. 🚀 **Deploy** - Build for production with `npm run build`

## 💡 Pro Tips

1. **Use Templates** - Click template buttons instead of typing everything
2. **JSON Validation** - The editor will show errors if JSON is invalid
3. **Debug Mode** - Expand the debug section to see exact API calls
4. **Dark Mode** - Automatically follows your system preference
5. **Responsive** - Works great on mobile and tablet too!

## 🎉 You're All Set!

The frontend is now running and ready to create approval requests.

**Quick Reference:**
- Start: `npm run dev`
- Build: `npm run build`
- Preview: `npm run preview`
- Lint: `npm run lint`

**Need Help?**
- Check [README.md](./README.md) for detailed documentation
- Review [EXAMPLES.md](./EXAMPLES.md) for more request examples
- See [CHECKLIST.md](./CHECKLIST.md) to verify everything works

Happy approving! 🚀

