import React, { useState, useEffect, useRef } from 'react';
import ApprovalForm from './components/ApprovalForm';
import WorkflowVisualization from './components/WorkflowVisualization';
import { approvalApi } from './services/api';
import { GenericApprovalRequest, WorkflowStageInfo, RequestTemplate } from './types/approval';
import { AlertCircle, CheckCircle, History, Slack } from 'lucide-react';

const requestTemplates: RequestTemplate[] = [
  {
    name: 'Deployment',
    request_type: 'deployment',
    message: 'Please approve production deployment',
    request_data: {
      service: 'api-gateway',
      version: 'v2.1.0',
      environment: 'production',
    },
  },
  {
    name: 'Access Request',
    request_type: 'access_request',
    message: 'Requesting access to production database',
    request_data: {
      resource: 'production-db',
      access_level: 'read-only',
      duration: '24 hours',
    },
  },
  {
    name: 'Code Review',
    request_type: 'code_review',
    message: 'Please review PR #123 for the new feature',
    request_data: {
      pr_number: '123',
      repository: 'backend-api',
      branch: 'feature/new-approval-system',
      lines_changed: 450,
    },
  },
];

const initialStages: WorkflowStageInfo[] = [
  {
    stage: 'created',
    label: 'Request Created',
    description: 'Approval request initialized',
    completed: false,
    active: false,
  },
  {
    stage: 'kafka_requested',
    label: 'Published to Kafka',
    description: 'Message sent to approval.requested topic',
    completed: false,
    active: false,
  },
  {
    stage: 'slack_sent',
    label: 'Sent to Slack DM',
    description: 'Approval message delivered to approver',
    completed: false,
    active: false,
  },
  {
    stage: 'pending_approval',
    label: 'Pending Approval',
    description: 'Waiting for approver action',
    completed: false,
    active: false,
  },
  {
    stage: 'approved',
    label: 'Approved/Rejected',
    description: 'Approver has made a decision',
    completed: false,
    active: false,
  },
  {
    stage: 'kafka_completed',
    label: 'Published to Kafka',
    description: 'Result sent to approval.completed topic',
    completed: false,
    active: false,
  },
];

function App() {
  const [isLoading, setIsLoading] = useState(false);
  const [stages, setStages] = useState<WorkflowStageInfo[]>(initialStages);
  const [requestId, setRequestId] = useState<string | undefined>();
  const [responseMessage, setResponseMessage] = useState<{
    type: 'success' | 'error';
    message: string;
  } | null>(null);
  const [showDebug, setShowDebug] = useState(false);
  const [lastRequest, setLastRequest] = useState<GenericApprovalRequest | null>(null);
  const [lastResponse, setLastResponse] = useState<any>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [wsConnected, setWsConnected] = useState(false);

  // WebSocket connection
  useEffect(() => {
    const wsUrl = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083').replace('http', 'ws') + '/ws';
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setWsConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log('WebSocket message received:', data);

        if (data.type === 'approval_update' && data.request_id === requestId) {
          // Update stage 4 (approved/rejected)
          updateStage(3, true, false); // Complete pending approval
          updateStage(4, true, false); // Mark as approved/rejected

          setResponseMessage({
            type: data.approved ? 'success' : 'error',
            message: data.approved
              ? `✅ Approval ${data.status}: ${data.message}`
              : `❌ Approval ${data.status}: ${data.message}`,
          });
        }

        if (data.type === 'action_update' && data.request_id === requestId) {
          // Update stage 5 (kafka_completed)
          updateStage(5, true, false);

          setResponseMessage({
            type: data.status === 'executed' ? 'success' : 'error',
            message: `🎉 Workflow complete! Action ${data.status}: ${data.message}`,
          });
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setWsConnected(false);
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
      setWsConnected(false);
    };

    wsRef.current = ws;

    return () => {
      ws.close();
    };
  }, [requestId]);

  const resetWorkflow = () => {
    setStages(initialStages);
    setRequestId(undefined);
    setResponseMessage(null);
    setShowDebug(false);
    setLastRequest(null);
    setLastResponse(null);
  };

  const updateStage = (stageIndex: number, completed: boolean, active: boolean) => {
    setStages((prev) =>
      prev.map((stage, index) => {
        if (index === stageIndex) {
          return { ...stage, completed, active };
        }
        return stage;
      })
    );
  };

  const handleSubmit = async (request: GenericApprovalRequest) => {
    setIsLoading(true);
    resetWorkflow();
    setLastRequest(request);

    try {
      updateStage(0, true, false);
      await new Promise((resolve) => setTimeout(resolve, 500));

      const response = await approvalApi.createApprovalRequest(request);
      setLastResponse(response);

      if (response.success && response.request_id) {
        setRequestId(response.request_id);
        setResponseMessage({
          type: 'success',
          message: response.message,
        });

        updateStage(0, true, false);
        updateStage(1, true, false);
        await new Promise((resolve) => setTimeout(resolve, 500));

        updateStage(2, true, false);
        await new Promise((resolve) => setTimeout(resolve, 500));

        updateStage(3, false, true);
      } else {
        setResponseMessage({
          type: 'error',
          message: response.error || 'Failed to create approval request',
        });
        setStages((prev) =>
          prev.map((stage, index) =>
            index === 0 ? { ...stage, completed: false, active: false, error: true } : stage
          )
        );
      }
    } catch (error) {
      setResponseMessage({
        type: 'error',
        message: 'An unexpected error occurred',
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 py-8 px-4">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="flex items-center justify-center mb-4">
            <Slack className="w-12 h-12 text-slack-purple mr-3" />
            <h1 className="text-4xl font-bold text-gray-900 dark:text-white">
              Slack Approval Workflow
            </h1>
          </div>
          <p className="text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
            Create and track approval requests through the complete Kafka-driven workflow.
            Submit requests, monitor their progress through Slack, and see real-time status updates.
          </p>
          {/* WebSocket Status Indicator */}
          <div className="mt-4 flex items-center justify-center">
            <div className={`flex items-center px-3 py-1 rounded-full text-sm ${
              wsConnected
                ? 'bg-green-100 dark:bg-green-900/20 text-green-700 dark:text-green-400'
                : 'bg-red-100 dark:bg-red-900/20 text-red-700 dark:text-red-400'
            }`}>
              <div className={`w-2 h-2 rounded-full mr-2 ${
                wsConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'
              }`}></div>
              {wsConnected ? 'Real-time updates connected' : 'Real-time updates disconnected'}
            </div>
          </div>
        </div>

        {/* Response Message */}
        {responseMessage && (
          <div className="mb-6">
            <div
              className={`p-4 rounded-lg flex items-start ${
                responseMessage.type === 'success'
                  ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800'
                  : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'
              }`}
            >
              {responseMessage.type === 'success' ? (
                <CheckCircle className="w-5 h-5 text-green-600 dark:text-green-400 mr-3 flex-shrink-0 mt-0.5" />
              ) : (
                <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400 mr-3 flex-shrink-0 mt-0.5" />
              )}
              <div className="flex-1">
                <p
                  className={`font-semibold ${
                    responseMessage.type === 'success'
                      ? 'text-green-800 dark:text-green-300'
                      : 'text-red-800 dark:text-red-300'
                  }`}
                >
                  {responseMessage.message}
                </p>
                {requestId && (
                  <p className="text-sm text-green-700 dark:text-green-400 mt-1">
                    Request ID: <span className="font-mono font-semibold">{requestId}</span>
                  </p>
                )}
              </div>
              <button
                onClick={() => setResponseMessage(null)}
                className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              >
                ✕
              </button>
            </div>
          </div>
        )}

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Left Column - Form */}
          <div>
            <ApprovalForm
              onSubmit={handleSubmit}
              isLoading={isLoading}
              templates={requestTemplates}
            />
          </div>

          {/* Right Column - Workflow Visualization */}
          <div>
            <WorkflowVisualization stages={stages} requestId={requestId} />
          </div>
        </div>

        {/* Debug Section */}
        {(lastRequest || lastResponse) && (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
            <button
              onClick={() => setShowDebug(!showDebug)}
              className="flex items-center justify-between w-full text-left"
            >
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white flex items-center">
                <History className="w-5 h-5 mr-2" />
                Request/Response Debug Info
              </h3>
              <span className="text-gray-500">{showDebug ? '▼' : '▶'}</span>
            </button>

            {showDebug && (
              <div className="mt-4 space-y-4">
                {lastRequest && (
                  <div>
                    <h4 className="font-semibold text-gray-700 dark:text-gray-300 mb-2">
                      Last Request:
                    </h4>
                    <pre className="bg-gray-50 dark:bg-gray-900 p-4 rounded-md overflow-x-auto text-sm">
                      {JSON.stringify(lastRequest, null, 2)}
                    </pre>
                  </div>
                )}

                {lastResponse && (
                  <div>
                    <h4 className="font-semibold text-gray-700 dark:text-gray-300 mb-2">
                      Last Response:
                    </h4>
                    <pre className="bg-gray-50 dark:bg-gray-900 p-4 rounded-md overflow-x-auto text-sm">
                      {JSON.stringify(lastResponse, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="mt-8 text-center text-sm text-gray-600 dark:text-gray-400">
          <p>
            API Endpoint: <span className="font-mono font-semibold">
              {import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083'}/api/v1/approval/generic
            </span>
          </p>
          <p className="mt-2">
            Workflow: Request Created → Kafka (approval.requested) → Slack DM → Pending Approval →
            Approved/Rejected → Kafka (approval.completed)
          </p>
        </div>
      </div>
    </div>
  );
}

export default App;

