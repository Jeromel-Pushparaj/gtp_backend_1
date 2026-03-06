import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Upload, FileText, Clock, CheckCircle, XCircle } from 'lucide-react';
import { api } from '../services/api';
import type { RunSummary } from '../types';

export default function Dashboard() {
  const [runs, setRuns] = useState<RunSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();

  useEffect(() => {
    loadRuns();
    const interval = setInterval(loadRuns, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, []);

  const loadRuns = async () => {
    try {
      const data = await api.listRuns();
      setRuns(data.runs);
    } catch (error) {
      console.error('Failed to load runs:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0]);
    }
  };

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      handleFile(e.target.files[0]);
    }
  };

  const handleFile = async (file: File) => {
    if (!file.name.endsWith('.yaml') && !file.name.endsWith('.yml') && !file.name.endsWith('.json')) {
      alert('Please upload a YAML or JSON file');
      return;
    }

    setUploading(true);
    try {
      const result = await api.uploadSpec(file, file.name, '1.0.0');
      navigate(`/workflow/${result.run_id}`);
    } catch (error) {
      console.error('Failed to upload spec:', error);
      alert('Failed to upload spec. Please try again.');
    } finally {
      setUploading(false);
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-5 h-5 text-success" />;
      case 'failed':
        return <XCircle className="w-5 h-5 text-error" />;
      case 'running':
        return <Clock className="w-5 h-5 text-warning animate-spin" />;
      default:
        return <Clock className="w-5 h-5 text-gray-400" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-success';
      case 'failed':
        return 'bg-error';
      case 'running':
        return 'bg-warning';
      default:
        return 'bg-gray-400';
    }
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return '-';
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    return `${minutes}m ${seconds % 60}s`;
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);
    
    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Upload Section */}
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-4">Upload API Specification</h2>
        <div
          className={`border-2 border-dashed rounded-lg p-12 text-center transition-colors ${
            dragActive ? 'border-primary bg-blue-50' : 'border-gray-300 hover:border-primary'
          }`}
          onDragEnter={handleDrag}
          onDragLeave={handleDrag}
          onDragOver={handleDrag}
          onDrop={handleDrop}
        >
          <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <p className="text-lg text-gray-700 mb-2">
            {uploading ? 'Uploading...' : 'Drag & drop your Swagger/OpenAPI file here'}
          </p>
          <p className="text-sm text-gray-500 mb-4">or</p>
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
            className="bg-primary text-white px-6 py-2 rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Browse Files
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".yaml,.yml,.json"
            onChange={handleFileInput}
            className="hidden"
          />
        </div>
      </div>

      {/* Recent Runs */}
      <div>
        <h2 className="text-2xl font-bold text-gray-900 mb-4">Recent Test Runs</h2>
        {loading ? (
          <div className="text-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          </div>
        ) : runs.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg border">
            <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-500">No test runs yet. Upload a spec to get started!</p>
          </div>
        ) : (
          <div className="space-y-4">
            {runs.map((run) => (
              <div
                key={run.id}
                className="bg-white rounded-lg border p-6 hover:shadow-md transition-shadow"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4 flex-1">
                    {getStatusIcon(run.status)}
                    <div className="flex-1">
                      <h3 className="text-lg font-semibold text-gray-900">
                        {run.spec_name} <span className="text-sm text-gray-500">v{run.spec_version}</span>
                      </h3>
                      <p className="text-sm text-gray-500">{formatTime(run.started_at)}</p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-6">
                    {run.status === 'running' && (
                      <div className="text-center">
                        <p className="text-sm text-gray-500">Phase {run.phase.replace('phase_', '')}</p>
                        <div className="w-32 bg-gray-200 rounded-full h-2 mt-1">
                          <div
                            className={`h-2 rounded-full ${getStatusColor(run.status)}`}
                            style={{ width: `${run.progress}%` }}
                          ></div>
                        </div>
                      </div>
                    )}
                    {run.status === 'completed' && (
                      <div className="text-center">
                        <p className="text-2xl font-bold text-gray-900">
                          {run.passed_tests}/{run.total_tests}
                        </p>
                        <p className="text-sm text-gray-500">Passed</p>
                      </div>
                    )}
                    <div className="text-right">
                      <p className="text-sm text-gray-500">Duration</p>
                      <p className="text-sm font-medium text-gray-900">{formatDuration(run.duration_ms)}</p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={() => navigate(`/workflow/${run.id}`)}
                        className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-blue-600 transition-colors text-sm"
                      >
                        View Workflow
                      </button>
                      {run.status === 'completed' && (
                        <button
                          onClick={() => navigate(`/report/${run.id}`)}
                          className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors text-sm"
                        >
                          View Report
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

