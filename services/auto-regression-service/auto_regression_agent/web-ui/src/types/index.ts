export interface RunSummary {
  id: string;
  spec_name: string;
  spec_version: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  phase: 'phase_1' | 'phase_2' | 'phase_3';
  progress: number;
  total_tests: number;
  passed_tests: number;
  failed_tests: number;
  started_at: string;
  completed_at?: string;
  duration_ms?: number;
}

export interface EndpointInfo {
  method: string;
  path: string;
  tests_count: number;
  passed_count: number;
  failed_count: number;
}

export interface AgentActivity {
  agent: string;
  status: 'pending' | 'active' | 'completed' | 'failed';
  message: string;
  details?: Record<string, any>;
  timestamp: string;
}

export interface SchemaDrift {
  endpoint: string;
  drift_type: string;
  field_path: string;
  old_value?: string;
  new_value?: string;
  confidence: number;
  auto_fixed: boolean;
}

export interface SecurityFinding {
  endpoint: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  category: string;
  description: string;
  payload: string;
  response?: string;
}

export interface RunDetails extends RunSummary {
  endpoints: EndpointInfo[];
  agent_activities: AgentActivity[];
  schema_drifts: SchemaDrift[];
  security_findings: SecurityFinding[];
}

export interface ValidationResult {
  type: string;
  expected: string;
  actual: string;
  passed: boolean;
  message?: string;
}

export interface TestResult {
  id: string;
  endpoint: string;
  method: string;
  test_type: 'smart_data' | 'mutation' | 'security' | 'performance';
  status: 'passed' | 'failed' | 'skipped';
  duration_ms: number;
  request: Record<string, any>;
  response: Record<string, any>;
  error?: string;
  validations: ValidationResult[];
}

export interface LogEntry {
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
  timestamp: string;
  agent?: string;
  details?: Record<string, any>;
}

export interface WebSocketMessage {
  type: 'connected' | 'workflow_status' | 'agent_activity' | 'log' | 'phase_complete';
  run_id: string;
  data?: any;
  timestamp: string;
}

