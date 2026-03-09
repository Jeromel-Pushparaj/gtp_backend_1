# Installation & Verification Checklist

Use this checklist to verify your frontend setup is complete and working.

## ✅ Pre-Installation Checklist

- [ ] Node.js 18+ installed (`node --version`)
- [ ] npm/yarn/pnpm available (`npm --version`)
- [ ] Backend service available at `http://localhost:8083`
- [ ] Backend health check passes: `curl http://localhost:8083/health`

## ✅ Installation Checklist

- [ ] Navigate to frontend directory: `cd frontend`
- [ ] Run setup script: `./setup.sh` OR `npm install`
- [ ] Verify `.env` file exists
- [ ] Verify `node_modules` directory created
- [ ] No installation errors in console

## ✅ File Structure Checklist

### Configuration Files
- [ ] `package.json` - Dependencies and scripts
- [ ] `tsconfig.json` - TypeScript configuration
- [ ] `vite.config.ts` - Vite configuration
- [ ] `tailwind.config.js` - Tailwind CSS configuration
- [ ] `.env` - Environment variables
- [ ] `.eslintrc.cjs` - ESLint configuration

### Source Files
- [ ] `src/main.tsx` - React entry point
- [ ] `src/App.tsx` - Main application component
- [ ] `src/index.css` - Global styles
- [ ] `src/vite-env.d.ts` - TypeScript environment types

### Components
- [ ] `src/components/ApprovalForm.tsx`
- [ ] `src/components/WorkflowVisualization.tsx`
- [ ] `src/components/StatusIndicator.tsx`

### Services & Types
- [ ] `src/services/api.ts` - API client
- [ ] `src/types/approval.ts` - TypeScript types

### Documentation
- [ ] `README.md` - Full documentation
- [ ] `QUICKSTART.md` - Quick start guide
- [ ] `EXAMPLES.md` - Example requests
- [ ] `ARCHITECTURE.md` - Architecture documentation
- [ ] `PROJECT_SUMMARY.md` - Project overview
- [ ] `CHECKLIST.md` - This file

## ✅ Development Server Checklist

- [ ] Start dev server: `npm run dev`
- [ ] Server starts without errors
- [ ] Browser opens automatically at `http://localhost:3000`
- [ ] No console errors in browser
- [ ] Page loads and displays correctly

## ✅ UI Component Checklist

### Header
- [ ] Slack icon visible
- [ ] "Slack Approval Workflow" title displayed
- [ ] Description text visible

### Approval Form (Left Column)
- [ ] "Create Approval Request" header visible
- [ ] Quick Templates section with 3 buttons visible
- [ ] All form fields render correctly:
  - [ ] Bot ID (optional)
  - [ ] Approver Name (required)
  - [ ] Requester Name (required)
  - [ ] Request Type dropdown
  - [ ] Message textarea
  - [ ] Request Data JSON editor
  - [ ] Use App DM checkbox
  - [ ] App Bot User ID (conditional)
- [ ] Submit button visible

### Workflow Visualization (Right Column)
- [ ] "Approval Workflow Status" header visible
- [ ] 6 workflow stages displayed
- [ ] All stages show inactive state initially
- [ ] Legend section visible with 4 status types

### Footer
- [ ] API endpoint displayed
- [ ] Workflow description visible

## ✅ Functionality Checklist

### Template Loading
- [ ] Click "Deployment" template
- [ ] Form fields auto-populate
- [ ] JSON editor shows formatted data
- [ ] Click "Access Request" template
- [ ] Form updates with new data
- [ ] Click "Code Review" template
- [ ] Form updates again

### Form Validation
- [ ] Clear all fields
- [ ] Click Submit
- [ ] Required field errors appear
- [ ] Error messages are red with icons
- [ ] Fill in Approver Name
- [ ] Error disappears for that field
- [ ] Fill in Requester Name
- [ ] Error disappears
- [ ] Fill in Message
- [ ] Error disappears
- [ ] Check "Use App DM"
- [ ] App Bot User ID field appears
- [ ] Submit without App Bot User ID
- [ ] Error appears for App Bot User ID

### JSON Editor
- [ ] Enter invalid JSON: `{invalid`
- [ ] Error message appears below editor
- [ ] Enter valid JSON: `{"test": "value"}`
- [ ] Error disappears
- [ ] Clear JSON editor
- [ ] No error (optional field)

### API Integration (Backend Running)
- [ ] Load "Deployment" template
- [ ] Fill in all required fields
- [ ] Click Submit
- [ ] Loading spinner appears on button
- [ ] Button is disabled during submission
- [ ] Success message appears (green)
- [ ] Request ID displayed
- [ ] Workflow stages update:
  - [ ] Stage 1: Created (green checkmark)
  - [ ] Stage 2: Kafka Requested (green checkmark)
  - [ ] Stage 3: Slack DM (green checkmark)
  - [ ] Stage 4: Pending (blue, in-progress)
  - [ ] Stages 5-6: Inactive

### Debug Section
- [ ] "Request/Response Debug Info" section visible
- [ ] Click to expand
- [ ] Last Request JSON displayed
- [ ] Last Response JSON displayed
- [ ] JSON is properly formatted
- [ ] Click to collapse
- [ ] Section collapses

### Error Handling (Backend Offline)
- [ ] Stop backend service
- [ ] Submit a request
- [ ] Error message appears (red)
- [ ] Error message is descriptive
- [ ] Workflow shows error state
- [ ] Can dismiss error message

## ✅ Responsive Design Checklist

- [ ] Resize browser to mobile width (< 768px)
- [ ] Form and workflow stack vertically
- [ ] All elements remain readable
- [ ] Buttons are touch-friendly
- [ ] No horizontal scrolling
- [ ] Resize to tablet width (768px - 1024px)
- [ ] Layout adjusts appropriately
- [ ] Resize to desktop width (> 1024px)
- [ ] Two-column layout displays

## ✅ Dark Mode Checklist

- [ ] Change system to dark mode
- [ ] Page background is dark
- [ ] Text is light colored
- [ ] Form inputs have dark background
- [ ] Borders are visible
- [ ] All text is readable
- [ ] Change back to light mode
- [ ] Everything reverts correctly

## ✅ Build Checklist

- [ ] Run build: `npm run build`
- [ ] Build completes without errors
- [ ] `dist/` directory created
- [ ] Files in `dist/` directory
- [ ] Run preview: `npm run preview`
- [ ] Preview server starts
- [ ] Application works in preview mode

## ✅ Code Quality Checklist

- [ ] Run linter: `npm run lint`
- [ ] No linting errors
- [ ] TypeScript compiles without errors
- [ ] All imports resolve correctly

## 🎯 Final Verification

- [ ] All checkboxes above are checked ✓
- [ ] Application runs smoothly
- [ ] No console errors
- [ ] Can create approval requests successfully
- [ ] Workflow visualization updates correctly
- [ ] Documentation is clear and helpful

## 🐛 If Something Fails

1. **Installation Issues**
   ```bash
   rm -rf node_modules package-lock.json
   npm install
   ```

2. **Build Issues**
   ```bash
   rm -rf node_modules/.vite dist
   npm run build
   ```

3. **Backend Connection Issues**
   - Check backend is running: `curl http://localhost:8083/health`
   - Verify `.env` has correct URL
   - Check browser console for CORS errors

4. **TypeScript Errors**
   - Ensure TypeScript version matches: `npm list typescript`
   - Clear cache: `rm -rf node_modules/.vite`

## ✅ Success!

If all items are checked, your frontend is fully functional! 🎉

Next steps:
1. Read the full [README.md](./README.md)
2. Try examples from [EXAMPLES.md](./EXAMPLES.md)
3. Review architecture in [ARCHITECTURE.md](./ARCHITECTURE.md)
4. Start building your approval workflows!

