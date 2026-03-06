import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Network, MessageSquare, Users, GitBranch, Clock, CheckCircle, AlertCircle } from 'lucide-react';
// import { api } from '../services/api'; // TODO: Uncomment when API is ready

interface AgentFeedback {
  id: string;
  from_agent: string;
  to_agent: string;
  feedback_type: string;
  message: string;
  priority: number;
  timestamp: string;
  acknowledged: boolean;
}

interface ConsensusDecision {
  id: string;
  decision_type: string;
  participating_agents: string[];
  votes: { [agent: string]: { option: string; confidence: number } };
  final_decision: string;
  confidence: number;
  timestamp: string;
}

interface AgentNode {
  id: string;
  name: string;
  status: 'active' | 'idle' | 'completed';
  messages_sent: number;
  messages_received: number;
}

export default function AgentCollaboration() {
  const { runId } = useParams<{ runId: string }>();
  const [feedbacks, setFeedbacks] = useState<AgentFeedback[]>([]);
  const [decisions, setDecisions] = useState<ConsensusDecision[]>([]);
  const [agents, setAgents] = useState<AgentNode[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'network' | 'timeline' | 'decisions'>('network');

  useEffect(() => {
    loadCollaborationData();
    const interval = setInterval(loadCollaborationData, 3000);
    return () => clearInterval(interval);
  }, [runId]);

  const loadCollaborationData = async () => {
    try {
      // TODO: Replace with actual API calls
      // const data = await api.getAgentCollaboration(runId!);
      
      // Mock data for now
      const mockFeedbacks: AgentFeedback[] = [
        {
          id: '1',
          from_agent: 'smart_data_generator',
          to_agent: 'mutation_generator',
          feedback_type: 'data_pattern',
          message: 'Found common pattern: email validation in 5 endpoints',
          priority: 2,
          timestamp: new Date(Date.now() - 120000).toISOString(),
          acknowledged: true,
        },
        {
          id: '2',
          from_agent: 'schema_drift_detector',
          to_agent: 'automated_fix_generator',
          feedback_type: 'drift_detected',
          message: 'Field type changed: user.age (string → number)',
          priority: 3,
          timestamp: new Date(Date.now() - 60000).toISOString(),
          acknowledged: true,
        },
        {
          id: '3',
          from_agent: 'security_generator',
          to_agent: 'test_executor',
          feedback_type: 'security_concern',
          message: 'SQL injection vulnerability detected in /api/search',
          priority: 5,
          timestamp: new Date(Date.now() - 30000).toISOString(),
          acknowledged: false,
        },
      ];

      const mockDecisions: ConsensusDecision[] = [
        {
          id: '1',
          decision_type: 'test_priority',
          participating_agents: ['smart_data_generator', 'security_generator', 'performance_generator'],
          votes: {
            smart_data_generator: { option: 'security_first', confidence: 0.85 },
            security_generator: { option: 'security_first', confidence: 0.95 },
            performance_generator: { option: 'balanced', confidence: 0.70 },
          },
          final_decision: 'security_first',
          confidence: 0.83,
          timestamp: new Date(Date.now() - 180000).toISOString(),
        },
      ];

      const mockAgents: AgentNode[] = [
        { id: '1', name: 'smart_data_generator', status: 'active', messages_sent: 5, messages_received: 2 },
        { id: '2', name: 'mutation_generator', status: 'active', messages_sent: 3, messages_received: 4 },
        { id: '3', name: 'security_generator', status: 'active', messages_sent: 8, messages_received: 1 },
        { id: '4', name: 'schema_drift_detector', status: 'completed', messages_sent: 4, messages_received: 0 },
        { id: '5', name: 'automated_fix_generator', status: 'active', messages_sent: 2, messages_received: 3 },
        { id: '6', name: 'test_executor', status: 'idle', messages_sent: 0, messages_received: 5 },
      ];

      setFeedbacks(mockFeedbacks);
      setDecisions(mockDecisions);
      setAgents(mockAgents);
    } catch (error) {
      console.error('Failed to load collaboration data:', error);
    }
  };

  const getAgentColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-100 border-green-500 text-green-800';
      case 'completed': return 'bg-blue-100 border-blue-500 text-blue-800';
      default: return 'bg-gray-100 border-gray-500 text-gray-800';
    }
  };

  const getPriorityColor = (priority: number) => {
    if (priority >= 4) return 'bg-red-100 text-red-800 border-red-200';
    if (priority >= 3) return 'bg-orange-100 text-orange-800 border-orange-200';
    if (priority >= 2) return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    return 'bg-blue-100 text-blue-800 border-blue-200';
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    if (diff < 60) return `${diff}s ago`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return date.toLocaleDateString();
  };

  const filteredFeedbacks = selectedAgent
    ? feedbacks.filter(f => f.from_agent === selectedAgent || f.to_agent === selectedAgent)
    : feedbacks;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Agent Collaboration</h1>
          <p className="text-gray-600">Real-time visualization of AI agent communication and consensus decisions</p>
        </div>

        {/* View Mode Tabs */}
        <div className="bg-white rounded-lg shadow-sm border mb-6">
          <div className="flex border-b">
            <button
              onClick={() => setViewMode('network')}
              className={`flex items-center gap-2 px-6 py-3 font-medium ${
                viewMode === 'network'
                  ? 'border-b-2 border-primary text-primary'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              <Network className="w-5 h-5" />
              Network View
            </button>
            <button
              onClick={() => setViewMode('timeline')}
              className={`flex items-center gap-2 px-6 py-3 font-medium ${
                viewMode === 'timeline'
                  ? 'border-b-2 border-primary text-primary'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              <Clock className="w-5 h-5" />
              Timeline
            </button>
            <button
              onClick={() => setViewMode('decisions')}
              className={`flex items-center gap-2 px-6 py-3 font-medium ${
                viewMode === 'decisions'
                  ? 'border-b-2 border-primary text-primary'
                  : 'text-gray-600 hover:text-gray-900'
              }`}
            >
              <Users className="w-5 h-5" />
              Consensus Decisions
            </button>
          </div>
        </div>

        {/* Network View */}
        {viewMode === 'network' && (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Agent Nodes */}
            <div className="lg:col-span-2 bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
                <Network className="w-5 h-5 text-primary" />
                Agent Network
              </h2>

              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {agents.map((agent) => (
                  <button
                    key={agent.id}
                    onClick={() => setSelectedAgent(selectedAgent === agent.name ? null : agent.name)}
                    className={`p-4 rounded-lg border-2 transition-all ${
                      selectedAgent === agent.name
                        ? 'ring-2 ring-primary ring-offset-2'
                        : ''
                    } ${getAgentColor(agent.status)}`}
                  >
                    <div className="font-medium text-sm mb-2">{agent.name.replace(/_/g, ' ')}</div>
                    <div className="flex justify-between text-xs">
                      <span>↑ {agent.messages_sent}</span>
                      <span>↓ {agent.messages_received}</span>
                    </div>
                  </button>
                ))}
              </div>

              <div className="mt-6 p-4 bg-gray-50 rounded-lg">
                <div className="text-sm text-gray-600 mb-2">Legend:</div>
                <div className="flex flex-wrap gap-4 text-xs">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded bg-green-500"></div>
                    <span>Active</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded bg-blue-500"></div>
                    <span>Completed</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded bg-gray-500"></div>
                    <span>Idle</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span>↑ Sent</span>
                    <span>↓ Received</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Communication Stats */}
            <div className="space-y-4">
              <div className="bg-white rounded-lg shadow-sm border p-6">
                <h3 className="font-semibold mb-4">Communication Stats</h3>
                <div className="space-y-3">
                  <div>
                    <div className="text-sm text-gray-600">Total Messages</div>
                    <div className="text-2xl font-bold text-primary">{feedbacks.length}</div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600">Active Agents</div>
                    <div className="text-2xl font-bold text-success">
                      {agents.filter(a => a.status === 'active').length}
                    </div>
                  </div>
                  <div>
                    <div className="text-sm text-gray-600">Consensus Decisions</div>
                    <div className="text-2xl font-bold text-info">{decisions.length}</div>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg shadow-sm border p-6">
                <h3 className="font-semibold mb-4">Most Active Agents</h3>
                <div className="space-y-2">
                  {agents
                    .sort((a, b) => (b.messages_sent + b.messages_received) - (a.messages_sent + a.messages_received))
                    .slice(0, 5)
                    .map((agent) => (
                      <div key={agent.id} className="flex justify-between items-center text-sm">
                        <span className="truncate">{agent.name.replace(/_/g, ' ')}</span>
                        <span className="font-medium">{agent.messages_sent + agent.messages_received}</span>
                      </div>
                    ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Timeline View */}
        {viewMode === 'timeline' && (
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
              <MessageSquare className="w-5 h-5 text-primary" />
              Communication Timeline
              {selectedAgent && (
                <span className="text-sm font-normal text-gray-600">
                  (filtered by {selectedAgent.replace(/_/g, ' ')})
                </span>
              )}
            </h2>

            <div className="space-y-4">
              {filteredFeedbacks.map((feedback) => (
                <div key={feedback.id} className="border-l-4 border-primary pl-4 py-2">
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <GitBranch className="w-4 h-4 text-gray-400" />
                      <span className="font-medium text-sm">
                        {feedback.from_agent.replace(/_/g, ' ')}
                      </span>
                      <span className="text-gray-400">→</span>
                      <span className="font-medium text-sm">
                        {feedback.to_agent.replace(/_/g, ' ')}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getPriorityColor(feedback.priority)}`}>
                        Priority {feedback.priority}
                      </span>
                      {feedback.acknowledged ? (
                        <CheckCircle className="w-4 h-4 text-success" />
                      ) : (
                        <AlertCircle className="w-4 h-4 text-warning" />
                      )}
                    </div>
                  </div>
                  <div className="text-sm text-gray-700 mb-1">{feedback.message}</div>
                  <div className="flex items-center gap-4 text-xs text-gray-500">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {formatTimestamp(feedback.timestamp)}
                    </span>
                    <span className="px-2 py-0.5 bg-gray-100 rounded">{feedback.feedback_type}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Consensus Decisions View */}
        {viewMode === 'decisions' && (
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
              <Users className="w-5 h-5 text-primary" />
              Consensus Decisions
            </h2>

            <div className="space-y-6">
              {decisions.map((decision) => (
                <div key={decision.id} className="border rounded-lg p-4">
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <h3 className="font-semibold text-lg">{decision.decision_type.replace(/_/g, ' ')}</h3>
                      <p className="text-sm text-gray-600">{formatTimestamp(decision.timestamp)}</p>
                    </div>
                    <div className="text-right">
                      <div className="text-sm text-gray-600">Confidence</div>
                      <div className="text-2xl font-bold text-primary">
                        {(decision.confidence * 100).toFixed(0)}%
                      </div>
                    </div>
                  </div>

                  <div className="mb-4">
                    <div className="text-sm font-medium text-gray-700 mb-2">Participating Agents:</div>
                    <div className="flex flex-wrap gap-2">
                      {decision.participating_agents.map((agent) => (
                        <span key={agent} className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm">
                          {agent.replace(/_/g, ' ')}
                        </span>
                      ))}
                    </div>
                  </div>

                  <div className="mb-4">
                    <div className="text-sm font-medium text-gray-700 mb-2">Votes:</div>
                    <div className="space-y-2">
                      {Object.entries(decision.votes).map(([agent, vote]) => (
                        <div key={agent} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                          <span className="text-sm">{agent.replace(/_/g, ' ')}</span>
                          <div className="flex items-center gap-3">
                            <span className="text-sm font-medium">{vote.option}</span>
                            <span className="text-xs text-gray-600">
                              {(vote.confidence * 100).toFixed(0)}% confident
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="p-3 bg-green-50 border border-green-200 rounded">
                    <div className="flex items-center gap-2">
                      <CheckCircle className="w-5 h-5 text-success" />
                      <span className="font-medium text-success">Final Decision:</span>
                      <span className="font-bold">{decision.final_decision}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

