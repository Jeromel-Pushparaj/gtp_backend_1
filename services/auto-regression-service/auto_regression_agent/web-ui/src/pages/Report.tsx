import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Download, CheckCircle, XCircle, AlertTriangle, ArrowLeft, Shield, Activity } from 'lucide-react';
import { api } from '../services/api';

export default function Report() {
  const { runId } = useParams<{ runId: string }>();
  const navigate = useNavigate();
  const [report, setReport] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!runId) return;
    loadReport();

    // Poll for report if workflow is in progress
    const interval = setInterval(() => {
      if (error && error.includes('in progress')) {
        loadReport();
      }
    }, 5000); // Poll every 5 seconds

    return () => clearInterval(interval);
  }, [runId, error]);

  const loadReport = async () => {
    if (!runId) return;
    try {
      const data = await api.getRunReport(runId);
      setReport(data);
      setError(null);
    } catch (err: any) {
      console.error('Failed to load report:', err);
      setError(err.message || 'Failed to load report');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = () => {
    if (!runId) return;
    api.downloadReport(runId, 'json');
  };

  const getTestStatusIcon = (status: string) => {
    switch (status) {
      case 'passed':
        return <CheckCircle className="w-5 h-5 text-success" />;
      case 'failed':
        return <XCircle className="w-5 h-5 text-error" />;
      default:
        return <AlertTriangle className="w-5 h-5 text-warning" />;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'high':
        return 'bg-orange-100 text-orange-800 border-orange-200';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      default:
        return 'bg-blue-100 text-blue-800 border-blue-200';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  // Show "in progress" message if workflow is still running
  if (error && error.includes('in progress')) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Activity className="w-12 h-12 text-primary mx-auto mb-4 animate-pulse" />
          <h2 className="text-xl font-semibold text-gray-900 mb-2">Workflow In Progress</h2>
          <p className="text-gray-600 mb-4">{error}</p>
          <p className="text-sm text-gray-500">This page will automatically refresh when the report is ready.</p>
          <button
            onClick={() => navigate('/')}
            className="mt-6 flex items-center mx-auto text-primary hover:text-blue-600"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  if (!report) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertTriangle className="w-12 h-12 text-error mx-auto mb-4" />
          <p className="text-gray-700">{error || 'Report not found'}</p>
          <button
            onClick={() => navigate('/')}
            className="mt-6 flex items-center mx-auto text-primary hover:text-blue-600"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  const { summary, test_results, schema_drifts, security_findings } = report;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8 flex items-center justify-between">
        <div>
          <button
            onClick={() => navigate('/')}
            className="flex items-center text-gray-600 hover:text-primary mb-4"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Dashboard
          </button>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            {report.spec_name} <span className="text-gray-500">v{report.spec_version}</span>
          </h1>
          <p className="text-gray-600">Test Execution Report</p>
        </div>
        <button
          onClick={handleDownload}
          className="flex items-center bg-primary text-white px-4 py-2 rounded-lg hover:bg-blue-600"
        >
          <Download className="w-4 h-4 mr-2" />
          Download Report
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="bg-white rounded-lg border p-6">
          <p className="text-sm text-gray-600 mb-2">Total Tests</p>
          <p className="text-3xl font-bold text-gray-900">{summary.total_tests}</p>
        </div>
        <div className="bg-white rounded-lg border p-6">
          <p className="text-sm text-gray-600 mb-2">Passed</p>
          <p className="text-3xl font-bold text-success">{summary.passed_tests}</p>
        </div>
        <div className="bg-white rounded-lg border p-6">
          <p className="text-sm text-gray-600 mb-2">Failed</p>
          <p className="text-3xl font-bold text-error">{summary.failed_tests}</p>
        </div>
        <div className="bg-white rounded-lg border p-6">
          <p className="text-sm text-gray-600 mb-2">Duration</p>
          <p className="text-3xl font-bold text-gray-900">
            {Math.floor(summary.duration_ms / 1000)}s
          </p>
        </div>
      </div>

      {/* Test Results */}
      <div className="bg-white rounded-lg border p-6 mb-8">
        <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
          <Activity className="w-6 h-6 mr-2" />
          Test Results
        </h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Endpoint
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Test Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Duration
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {test_results.map((test: any, index: number) => (
                <tr key={index} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      {getTestStatusIcon(test.status)}
                      <span className="ml-2 text-sm font-medium text-gray-900 capitalize">{test.status}</span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-sm font-mono text-gray-900">
                      {test.method} {test.endpoint}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="text-sm text-gray-600 capitalize">{test.test_type.replace('_', ' ')}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600">
                    {test.duration_ms}ms
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Schema Drifts */}
      {schema_drifts && schema_drifts.length > 0 && (
        <div className="bg-white rounded-lg border p-6 mb-8">
          <h2 className="text-xl font-bold text-gray-900 mb-4">Schema Drifts Detected</h2>
          <div className="space-y-4">
            {schema_drifts.map((drift: any, index: number) => (
              <div key={index} className="border-l-4 border-warning pl-4 py-2">
                <div className="flex items-center justify-between mb-1">
                  <span className="font-semibold text-gray-900">{drift.endpoint}</span>
                  {drift.auto_fixed && (
                    <span className="text-xs bg-success text-white px-2 py-1 rounded">Auto-Fixed</span>
                  )}
                </div>
                <p className="text-sm text-gray-700">
                  {drift.drift_type.replace('_', ' ').toUpperCase()}: {drift.field_path}
                </p>
                {drift.new_value && (
                  <p className="text-sm text-gray-600 mt-1">New value: {drift.new_value}</p>
                )}
                <p className="text-xs text-gray-500 mt-1">Confidence: {(drift.confidence * 100).toFixed(0)}%</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Security Findings */}
      {security_findings && security_findings.length > 0 && (
        <div className="bg-white rounded-lg border p-6">
          <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center">
            <Shield className="w-6 h-6 mr-2" />
            Security Findings
          </h2>
          <div className="space-y-4">
            {security_findings.map((finding: any, index: number) => (
              <div key={index} className={`border rounded-lg p-4 ${getSeverityColor(finding.severity)}`}>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-semibold">{finding.endpoint}</span>
                  <span className="text-xs font-bold uppercase">{finding.severity}</span>
                </div>
                <p className="text-sm font-medium mb-1">{finding.category.replace('_', ' ').toUpperCase()}</p>
                <p className="text-sm mb-2">{finding.description}</p>
                <p className="text-xs font-mono bg-white bg-opacity-50 p-2 rounded">
                  Payload: {finding.payload}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

