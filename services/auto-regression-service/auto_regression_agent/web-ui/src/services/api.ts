import type { RunSummary, RunDetails, LogEntry } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

export const api = {
  // Upload spec file
  async uploadSpec(file: File, name: string, _version: string): Promise<{ run_id: string; workflow_id: string }> {
    const formData = new FormData();
    formData.append('spec', file); // Backend expects 'spec' field name
    formData.append('name', name);
    formData.append('service', 'default'); // Required by backend
    formData.append('team_id', 'default-team'); // Required by backend
    formData.append('run_mode', 'full'); // Optional: smoke, full, nightly

    const response = await fetch(`${API_BASE_URL}/specs`, {
      method: 'POST',
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to upload spec');
    }

    const data = await response.json();
    // Backend returns workflow_id, but UI expects run_id
    return {
      run_id: data.workflow_id || data.spec_id,
      workflow_id: data.workflow_id,
    };
  },

  // List all runs
  async listRuns(page = 1, pageSize = 20, status?: string): Promise<{
    runs: RunSummary[];
    page: number;
    page_size: number;
    total: number;
    has_more: boolean;
  }> {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
    });

    if (status) {
      params.append('status', status);
    }

    const response = await fetch(`${API_BASE_URL}/runs?${params}`);

    if (!response.ok) {
      throw new Error('Failed to fetch runs');
    }

    return response.json();
  },

  // Get run details
  async getRun(runId: string): Promise<RunDetails> {
    const response = await fetch(`${API_BASE_URL}/runs/${runId}`);

    if (!response.ok) {
      throw new Error('Failed to fetch run details');
    }

    return response.json();
  },

  // Get run report
  async getRunReport(runId: string): Promise<any> {
    const response = await fetch(`${API_BASE_URL}/runs/${runId}/report`);

    // Handle 202 Accepted - workflow still in progress
    if (response.status === 202) {
      const data = await response.json();
      throw new Error(data.message || 'Workflow is still in progress. Please wait for completion.');
    }

    if (!response.ok) {
      throw new Error('Failed to fetch run report');
    }

    return response.json();
  },

  // Get run logs
  async getRunLogs(runId: string): Promise<{ run_id: string; logs: LogEntry[]; total: number }> {
    const response = await fetch(`${API_BASE_URL}/runs/${runId}/logs`);

    if (!response.ok) {
      throw new Error('Failed to fetch run logs');
    }

    return response.json();
  },

  // Download report
  downloadReport(runId: string, format = 'json') {
    window.open(`${API_BASE_URL}/runs/${runId}/download?format=${format}`, '_blank');
  },

  // Get test cases
  async getTestCases(runId: string): Promise<any> {
    const response = await fetch(`${API_BASE_URL}/runs/${runId}/test-cases`);

    if (!response.ok) {
      throw new Error('Failed to fetch test cases');
    }

    return response.json();
  },
};

// WebSocket connection
export function connectWebSocket(runId: string, onMessage: (message: any) => void): WebSocket {
  const wsUrl = (API_BASE_URL.replace('http', 'ws')) + `/ws/runs/${runId}`;
  const ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log('WebSocket connected for run:', runId);
  };

  ws.onmessage = (event) => {
    try {
      const message = JSON.parse(event.data);
      onMessage(message);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  ws.onclose = () => {
    console.log('WebSocket disconnected for run:', runId);
  };

  return ws;
}

