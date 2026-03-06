import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import WorkflowVisualization from './pages/WorkflowVisualization';
import Report from './pages/Report';
import AgentCollaboration from './pages/AgentCollaboration';
import TestCases from './pages/TestCases';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-50">
        <nav className="bg-white shadow-sm border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between h-16">
              <div className="flex items-center">
                <h1 className="text-2xl font-bold text-primary">
                  🤖 AI Agent Test Platform
                </h1>
              </div>
              <div className="flex items-center space-x-4">
                <a href="/" className="text-gray-700 hover:text-primary px-3 py-2 rounded-md text-sm font-medium">
                  Dashboard
                </a>
                <a href="https://github.com" target="_blank" rel="noopener noreferrer" className="text-gray-700 hover:text-primary px-3 py-2 rounded-md text-sm font-medium">
                  Docs
                </a>
              </div>
            </div>
          </div>
        </nav>

        <main>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/workflow/:runId" element={<WorkflowVisualization />} />
            <Route path="/collaboration/:runId" element={<AgentCollaboration />} />
            <Route path="/test-cases/:runId" element={<TestCases />} />
            <Route path="/report/:runId" element={<Report />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
