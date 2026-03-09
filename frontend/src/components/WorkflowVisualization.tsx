import React from 'react';
import { WorkflowStageInfo } from '../types/approval';
import StatusIndicator from './StatusIndicator';
import { ArrowDown } from 'lucide-react';

interface WorkflowVisualizationProps {
  stages: WorkflowStageInfo[];
  requestId?: string;
}

const WorkflowVisualization: React.FC<WorkflowVisualizationProps> = ({ stages, requestId }) => {
  const getStatusType = (stage: WorkflowStageInfo): 'pending' | 'in-progress' | 'completed' | 'rejected' | 'inactive' => {
    if (stage.error) return 'rejected';
    if (stage.completed) return 'completed';
    if (stage.active) return 'in-progress';
    return 'inactive';
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Approval Workflow Status
        </h2>
        {requestId && (
          <p className="text-sm text-gray-600 dark:text-gray-400">
            Request ID: <span className="font-mono font-semibold">{requestId}</span>
          </p>
        )}
      </div>

      <div className="space-y-4">
        {stages.map((stage, index) => (
          <React.Fragment key={stage.stage}>
            <StatusIndicator
              status={getStatusType(stage)}
              label={stage.label}
              description={stage.description}
            />
            {index < stages.length - 1 && (
              <div className="flex justify-center py-2">
                <ArrowDown className="w-5 h-5 text-gray-400" />
              </div>
            )}
          </React.Fragment>
        ))}
      </div>

      <div className="mt-6 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
        <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Legend</h3>
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 rounded-full bg-slack-green"></div>
            <span className="text-gray-700 dark:text-gray-300">Completed</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 rounded-full bg-slack-blue"></div>
            <span className="text-gray-700 dark:text-gray-300">In Progress</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 rounded-full bg-slack-yellow"></div>
            <span className="text-gray-700 dark:text-gray-300">Pending</span>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 rounded-full bg-slack-red"></div>
            <span className="text-gray-700 dark:text-gray-300">Rejected</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WorkflowVisualization;

