import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { CheckCircle, Clock, AlertCircle, ArrowRight, Bot, FileText, Network, List } from 'lucide-react';
import { api, connectWebSocket } from '../services/api';
import type { RunDetails, AgentActivity, LogEntry, WebSocketMessage } from '../types';

export default function WorkflowVisualization() {
  const { runId } = useParams<{ runId: string }>();
  const navigate = useNavigate();
  const [run, setRun] = useState<RunDetails | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!runId) return;

    // Load initial data
    loadRunDetails();
    loadLogs();

    // Connect WebSocket for real-time updates
    const ws = connectWebSocket(runId, handleWebSocketMessage);

    return () => {
      ws.close();
    };
  }, [runId]);

  const loadRunDetails = async () => {
    if (!runId) return;
    try {
      const data = await api.getRun(runId);
      setRun(data);
    } catch (error) {
      console.error('Failed to load run details:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadLogs = async () => {
    if (!runId) return;
    try {
      const data = await api.getRunLogs(runId);
      setLogs(data.logs);
    } catch (error) {
      console.error('Failed to load logs:', error);
    }
  };

  const handleWebSocketMessage = (message: WebSocketMessage) => {
    console.log('WebSocket message:', message);

    switch (message.type) {
      case 'workflow_status':
        if (run) {
          setRun({
            ...run,
            phase: message.data.phase,
            status: message.data.status,
            progress: message.data.progress,
          });
        }
        break;

      case 'agent_activity':
        if (run) {
          const newActivity: AgentActivity = {
            agent: message.data.agent,
            status: message.data.status,
            message: message.data.message,
            details: message.data.details,
            timestamp: message.timestamp,
          };
          setRun({
            ...run,
            agent_activities: [newActivity, ...run.agent_activities].slice(0, 10),
          });
        }
        break;

      case 'log':
        const newLog: LogEntry = {
          level: message.data.level,
          message: message.data.message,
          timestamp: message.timestamp,
          agent: message.data.agent,
          details: message.data.details,
        };
        setLogs((prev) => [newLog, ...prev].slice(0, 100));
        break;

      case 'phase_complete':
        if (run && message.data.phase === 'phase_3') {
          // Workflow complete, navigate to report
          setTimeout(() => {
            navigate(`/report/${runId}`);
          }, 2000);
        }
        break;
    }
  };

  const getPhaseStatus = (phaseNum: number) => {
    if (!run) return 'pending';
    const currentPhase = parseInt(run.phase.replace('phase_', ''));
    if (phaseNum < currentPhase) return 'completed';
    if (phaseNum === currentPhase) return run.status === 'running' ? 'in_progress' : run.status;
    return 'pending';
  };

  const getPhaseIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-8 h-8 text-success" />;
      case 'in_progress':
        return <Clock className="w-8 h-8 text-warning animate-spin" />;
      case 'failed':
        return <AlertCircle className="w-8 h-8 text-error" />;
      default:
        return <Clock className="w-8 h-8 text-gray-300" />;
    }
  };

  const getAgentStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-success text-white';
      case 'active':
        return 'bg-warning text-white';
      case 'failed':
        return 'bg-error text-white';
      default:
        return 'bg-gray-300 text-gray-700';
    }
  };

  const getLogLevelColor = (level: string) => {
    switch (level) {
      case 'error':
        return 'text-error';
      case 'warn':
        return 'text-warning';
      case 'info':
        return 'text-info';
      default:
        return 'text-gray-600';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!run) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertCircle className="w-12 h-12 text-error mx-auto mb-4" />
          <p className="text-gray-700">Run not found</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8 flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            {run.spec_name} <span className="text-gray-500">v{run.spec_version}</span>
          </h1>
          <p className="text-gray-600">Run ID: {run.id}</p>
        </div>
        <div className="flex gap-3">
          <Link
            to={`/test-cases/${runId}`}
            className="flex items-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
          >
            <List className="w-5 h-5" />
            View Test Cases
          </Link>
          <Link
            to={`/collaboration/${runId}`}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            <Network className="w-5 h-5" />
            View Agent Collaboration
          </Link>
          <Link
            to={`/report/${runId}`}
            className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
          >
            <FileText className="w-5 h-5" />
            View Report
          </Link>
        </div>
      </div>

      {/* Workflow Pipeline */}
      <div className="bg-white rounded-lg border p-8 mb-8">
        <h2 className="text-xl font-bold text-gray-900 mb-6">Workflow Progress</h2>
        <div className="flex items-center justify-between">
          {/* Phase 1 */}
          <div className="flex-1 text-center">
            <div className="flex justify-center mb-2">{getPhaseIcon(getPhaseStatus(1))}</div>
            <p className="font-semibold text-gray-900">Phase 1</p>
            <p className="text-sm text-gray-600">Spec Analysis</p>
          </div>

          <ArrowRight className="w-6 h-6 text-gray-400 mx-4" />

          {/* Phase 2 */}
          <div className="flex-1 text-center">
            <div className="flex justify-center mb-2">{getPhaseIcon(getPhaseStatus(2))}</div>
            <p className="font-semibold text-gray-900">Phase 2</p>
            <p className="text-sm text-gray-600">Test Generation</p>
          </div>

          <ArrowRight className="w-6 h-6 text-gray-400 mx-4" />

          {/* Phase 3 */}
          <div className="flex-1 text-center">
            <div className="flex justify-center mb-2">{getPhaseIcon(getPhaseStatus(3))}</div>
            <p className="font-semibold text-gray-900">Phase 3</p>
            <p className="text-sm text-gray-600">Test Execution</p>
          </div>
        </div>

        {/* Progress Bar */}
        <div className="mt-6">
          <div className="flex justify-between text-sm text-gray-600 mb-2">
            <span>Overall Progress</span>
            <span>{run.progress}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-3">
            <div
              className="bg-primary h-3 rounded-full transition-all duration-500"
              style={{ width: `${run.progress}%` }}
            ></div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* AI Agents Activity */}
        <div className="bg-white rounded-lg border p-6">
          <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
            <Bot className="w-6 h-6 mr-2" />
            AI Agents Activity
          </h2>
          <div className="space-y-4 max-h-96 overflow-y-auto">
            {run.agent_activities.length === 0 ? (
              <p className="text-gray-500 text-center py-8">No agent activity yet</p>
            ) : (
              run.agent_activities.map((activity, index) => (
                <div key={index} className="border-l-4 border-primary pl-4 py-2">
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-semibold text-gray-900">{activity.agent.replace(/_/g, ' ').toUpperCase()}</span>
                    <span className={`text-xs px-2 py-1 rounded ${getAgentStatusColor(activity.status)}`}>
                      {activity.status}
                    </span>
                  </div>
                  <p className="text-sm text-gray-700">{activity.message}</p>
                  {activity.details && (
                    <p className="text-xs text-gray-500 mt-1">
                      {JSON.stringify(activity.details)}
                    </p>
                  )}
                  <p className="text-xs text-gray-400 mt-1">{new Date(activity.timestamp).toLocaleTimeString()}</p>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Live Logs */}
        <div className="bg-white rounded-lg border p-6">
          <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
            <FileText className="w-6 h-6 mr-2" />
            Live Logs
          </h2>
          <div className="bg-gray-900 rounded p-4 max-h-96 overflow-y-auto font-mono text-sm">
            {logs.length === 0 ? (
              <p className="text-gray-400">No logs yet</p>
            ) : (
              logs.map((log, index) => (
                <div key={index} className="mb-2">
                  <span className="text-gray-500">[{new Date(log.timestamp).toLocaleTimeString()}]</span>{' '}
                  <span className={getLogLevelColor(log.level)}>{log.level.toUpperCase()}</span>{' '}
                  {log.agent && <span className="text-blue-400">{log.agent}:</span>}{' '}
                  <span className="text-gray-300">{log.message}</span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

