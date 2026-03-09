import React from 'react';
import { CheckCircle2, Circle, XCircle, Loader2 } from 'lucide-react';

interface StatusIndicatorProps {
  status: 'pending' | 'in-progress' | 'completed' | 'rejected' | 'inactive';
  label: string;
  description?: string;
}

const StatusIndicator: React.FC<StatusIndicatorProps> = ({ status, label, description }) => {
  const getIcon = () => {
    switch (status) {
      case 'completed':
        return <CheckCircle2 className="w-6 h-6 text-slack-green" />;
      case 'in-progress':
        return <Loader2 className="w-6 h-6 text-slack-blue animate-spin" />;
      case 'rejected':
        return <XCircle className="w-6 h-6 text-slack-red" />;
      case 'pending':
        return <Circle className="w-6 h-6 text-slack-yellow" />;
      default:
        return <Circle className="w-6 h-6 text-gray-300" />;
    }
  };

  const getTextColor = () => {
    switch (status) {
      case 'completed':
        return 'text-slack-green';
      case 'in-progress':
        return 'text-slack-blue';
      case 'rejected':
        return 'text-slack-red';
      case 'pending':
        return 'text-slack-yellow';
      default:
        return 'text-gray-400';
    }
  };

  return (
    <div className="flex items-start space-x-3">
      <div className="flex-shrink-0 mt-1">{getIcon()}</div>
      <div className="flex-1">
        <h4 className={`font-semibold ${getTextColor()}`}>{label}</h4>
        {description && (
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{description}</p>
        )}
      </div>
    </div>
  );
};

export default StatusIndicator;

