export type RequestType = 'deployment' | 'access_request' | 'code_review' | 'other';

export type ApprovalStatus = 'pending' | 'approved' | 'rejected';

export type WorkflowStage = 
  | 'created'
  | 'kafka_requested'
  | 'slack_sent'
  | 'pending_approval'
  | 'approved'
  | 'rejected'
  | 'kafka_completed';

export interface GenericApprovalRequest {
  bot_id?: string;
  approver_id?: string;
  approver_name?: string;
  requester_id?: string;
  requester_name?: string;
  request_type: RequestType;
  message: string;
  request_data?: Record<string, any>;
  use_app_dm: boolean;
  app_bot_user_id?: string;
}

export interface ApprovalResponse {
  success: boolean;
  message: string;
  request_id?: string;
  error?: string;
}

export interface ApprovalRequestHistory {
  id: string;
  request_type: RequestType;
  approver_name: string;
  requester_name: string;
  message: string;
  status: ApprovalStatus;
  created_at: string;
  request_data?: Record<string, any>;
}

export interface WorkflowStageInfo {
  stage: WorkflowStage;
  label: string;
  description: string;
  completed: boolean;
  active: boolean;
  error?: boolean;
}

export interface RequestTemplate {
  name: string;
  request_type: RequestType;
  message: string;
  request_data: Record<string, any>;
}

