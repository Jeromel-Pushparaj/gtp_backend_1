import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  FileText,
  CheckCircle,
  Clock,
  AlertTriangle,
  Code,
  Shield,
  Zap,
  Database,
  ChevronDown,
  ChevronRight,
  ArrowLeft
} from 'lucide-react';
import { api } from '../services/api';

interface TestCase {
  id: string;
  endpoint: string;
  method: string;
  test_type: string;
  category: string;
  description: string;
  payload: any;
  expected_status: number;
  validations: string[];
  generated_by: string;
  created_at: string;
}

interface TestCasesResponse {
  run_id: string;
  spec_name: string;
  total_test_cases: number;
  test_cases_by_type: {
    smart_data: number;
    mutation: number;
    security: number;
    performance: number;
  };
  test_cases: TestCase[];
}

export default function TestCases() {
  const { runId } = useParams<{ runId: string }>();
  const [data, setData] = useState<TestCasesResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [expandedTests, setExpandedTests] = useState<Set<string>>(new Set());
  const [filterType, setFilterType] = useState<string>('all');
  const [filterEndpoint, setFilterEndpoint] = useState<string>('all');

  useEffect(() => {
    loadTestCases();
  }, [runId]);

  const loadTestCases = async () => {
    if (!runId) return;
    try {
      const response = await api.getTestCases(runId);
      setData(response);
    } catch (error) {
      console.error('Failed to load test cases:', error);
    } finally {
      setLoading(false);
    }
  };

  const toggleExpand = (testId: string) => {
    const newExpanded = new Set(expandedTests);
    if (newExpanded.has(testId)) {
      newExpanded.delete(testId);
    } else {
      newExpanded.add(testId);
    }
    setExpandedTests(newExpanded);
  };

  const getTestTypeIcon = (type: string) => {
    switch (type) {
      case 'smart_data':
        return <Database className="w-5 h-5 text-blue-600" />;
      case 'mutation':
        return <Zap className="w-5 h-5 text-yellow-600" />;
      case 'security':
        return <Shield className="w-5 h-5 text-red-600" />;
      case 'performance':
        return <Clock className="w-5 h-5 text-green-600" />;
      default:
        return <FileText className="w-5 h-5 text-gray-600" />;
    }
  };

  const getTestTypeColor = (type: string) => {
    switch (type) {
      case 'smart_data':
        return 'bg-blue-100 text-blue-800 border-blue-300';
      case 'mutation':
        return 'bg-yellow-100 text-yellow-800 border-yellow-300';
      case 'security':
        return 'bg-red-100 text-red-800 border-red-300';
      case 'performance':
        return 'bg-green-100 text-green-800 border-green-300';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

  const getMethodColor = (method: string) => {
    switch (method) {
      case 'GET':
        return 'bg-green-100 text-green-800';
      case 'POST':
        return 'bg-blue-100 text-blue-800';
      case 'PUT':
        return 'bg-yellow-100 text-yellow-800';
      case 'DELETE':
        return 'bg-red-100 text-red-800';
      case 'PATCH':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertTriangle className="w-12 h-12 text-error mx-auto mb-4" />
          <p className="text-gray-700">Failed to load test cases</p>
        </div>
      </div>
    );
  }

  const filteredTests = data.test_cases.filter(test => {
    if (filterType !== 'all' && test.test_type !== filterType) return false;
    if (filterEndpoint !== 'all' && test.endpoint !== filterEndpoint) return false;
    return true;
  });

  const uniqueEndpoints = Array.from(new Set(data.test_cases.map(t => t.endpoint)));

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8">
        <Link
          to={`/workflow/${runId}`}
          className="inline-flex items-center gap-2 text-primary hover:text-blue-600 mb-4"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Workflow
        </Link>
        <h1 className="text-3xl font-bold text-gray-900 mb-2">
          Generated Test Cases
        </h1>
        <p className="text-gray-600">{data.spec_name}</p>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4 mb-8">
        <div className="bg-white rounded-lg border p-4">
          <p className="text-sm text-gray-500 mb-1">Total Tests</p>
          <p className="text-2xl font-bold text-gray-900">{data.total_test_cases}</p>
        </div>
        <div className="bg-blue-50 rounded-lg border border-blue-200 p-4">
          <div className="flex items-center gap-2 mb-1">
            <Database className="w-4 h-4 text-blue-600" />
            <p className="text-sm text-blue-600">Smart Data</p>
          </div>
          <p className="text-2xl font-bold text-blue-900">{data.test_cases_by_type.smart_data}</p>
        </div>
        <div className="bg-yellow-50 rounded-lg border border-yellow-200 p-4">
          <div className="flex items-center gap-2 mb-1">
            <Zap className="w-4 h-4 text-yellow-600" />
            <p className="text-sm text-yellow-600">Mutation</p>
          </div>
          <p className="text-2xl font-bold text-yellow-900">{data.test_cases_by_type.mutation}</p>
        </div>
        <div className="bg-red-50 rounded-lg border border-red-200 p-4">
          <div className="flex items-center gap-2 mb-1">
            <Shield className="w-4 h-4 text-red-600" />
            <p className="text-sm text-red-600">Security</p>
          </div>
          <p className="text-2xl font-bold text-red-900">{data.test_cases_by_type.security}</p>
        </div>
        <div className="bg-green-50 rounded-lg border border-green-200 p-4">
          <div className="flex items-center gap-2 mb-1">
            <Clock className="w-4 h-4 text-green-600" />
            <p className="text-sm text-green-600">Performance</p>
          </div>
          <p className="text-2xl font-bold text-green-900">{data.test_cases_by_type.performance}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg border p-4 mb-6">
        <div className="flex gap-4">
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-2">Filter by Type</label>
            <select
              value={filterType}
              onChange={(e) => setFilterType(e.target.value)}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="all">All Types ({data.total_test_cases})</option>
              <option value="smart_data">Smart Data ({data.test_cases_by_type.smart_data})</option>
              <option value="mutation">Mutation ({data.test_cases_by_type.mutation})</option>
              <option value="security">Security ({data.test_cases_by_type.security})</option>
              <option value="performance">Performance ({data.test_cases_by_type.performance})</option>
            </select>
          </div>
          <div className="flex-1">
            <label className="block text-sm font-medium text-gray-700 mb-2">Filter by Endpoint</label>
            <select
              value={filterEndpoint}
              onChange={(e) => setFilterEndpoint(e.target.value)}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="all">All Endpoints ({uniqueEndpoints.length})</option>
              {uniqueEndpoints.map(endpoint => (
                <option key={endpoint} value={endpoint}>{endpoint}</option>
              ))}
            </select>
          </div>
        </div>
        <p className="text-sm text-gray-500 mt-2">
          Showing {filteredTests.length} of {data.total_test_cases} test cases
        </p>
      </div>

      {/* Test Cases List */}
      <div className="space-y-4">
        {filteredTests.map((test) => (
          <div key={test.id} className="bg-white rounded-lg border">
            {/* Test Header */}
            <div
              onClick={() => toggleExpand(test.id)}
              className="p-4 cursor-pointer hover:bg-gray-50 transition-colors"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4 flex-1">
                  {expandedTests.has(test.id) ? (
                    <ChevronDown className="w-5 h-5 text-gray-400" />
                  ) : (
                    <ChevronRight className="w-5 h-5 text-gray-400" />
                  )}
                  {getTestTypeIcon(test.test_type)}
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${getMethodColor(test.method)}`}>
                        {test.method}
                      </span>
                      <span className="font-mono text-sm text-gray-900">{test.endpoint}</span>
                    </div>
                    <p className="text-sm text-gray-600">{test.description}</p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`px-3 py-1 rounded-full text-xs font-medium border ${getTestTypeColor(test.test_type)}`}>
                    {test.test_type.replace('_', ' ').toUpperCase()}
                  </span>
                  <span className="text-xs text-gray-500">by {test.generated_by}</span>
                </div>
              </div>
            </div>

            {/* Expanded Details */}
            {expandedTests.has(test.id) && (
              <div className="border-t bg-gray-50 p-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* Request Payload */}
                  <div>
                    <h4 className="text-sm font-semibold text-gray-900 mb-2 flex items-center gap-2">
                      <Code className="w-4 h-4" />
                      Request Payload
                    </h4>
                    <pre className="bg-gray-900 text-gray-100 p-3 rounded-lg text-xs overflow-x-auto">
                      {JSON.stringify(test.payload, null, 2)}
                    </pre>
                  </div>

                  {/* Test Details */}
                  <div className="space-y-4">
                    <div>
                      <h4 className="text-sm font-semibold text-gray-900 mb-2">Expected Response</h4>
                      <p className="text-sm text-gray-600">
                        Status Code: <span className="font-mono font-medium">{test.expected_status}</span>
                      </p>
                    </div>

                    <div>
                      <h4 className="text-sm font-semibold text-gray-900 mb-2">Validations</h4>
                      <ul className="space-y-1">
                        {test.validations.map((validation, idx) => (
                          <li key={idx} className="text-sm text-gray-600 flex items-start gap-2">
                            <CheckCircle className="w-4 h-4 text-green-600 mt-0.5 flex-shrink-0" />
                            <span>{validation}</span>
                          </li>
                        ))}
                      </ul>
                    </div>

                    <div>
                      <h4 className="text-sm font-semibold text-gray-900 mb-2">Metadata</h4>
                      <div className="text-sm text-gray-600 space-y-1">
                        <p>Category: <span className="font-medium">{test.category}</span></p>
                        <p>Test ID: <span className="font-mono text-xs">{test.id}</span></p>
                        <p>Created: <span className="font-medium">{new Date(test.created_at).toLocaleString()}</span></p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {filteredTests.length === 0 && (
        <div className="text-center py-12">
          <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-600">No test cases match the selected filters</p>
        </div>
      )}
    </div>
  );
}


