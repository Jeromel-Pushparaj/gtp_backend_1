import React, { useState } from 'react';
import { GenericApprovalRequest, RequestType, RequestTemplate } from '../types/approval';
import { Send, AlertCircle, CheckCircle, FileJson } from 'lucide-react';

interface ApprovalFormProps {
  onSubmit: (request: GenericApprovalRequest) => Promise<void>;
  isLoading: boolean;
  templates: RequestTemplate[];
}

const ApprovalForm: React.FC<ApprovalFormProps> = ({ onSubmit, isLoading, templates }) => {
  const [formData, setFormData] = useState<GenericApprovalRequest>({
    bot_id: '',
    approver_name: '',
    requester_name: '',
    request_type: 'deployment',
    message: '',
    request_data: undefined,
    use_app_dm: true,
    app_bot_user_id: '',
  });

  const [requestDataJson, setRequestDataJson] = useState('');
  const [jsonError, setJsonError] = useState('');
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value, type } = e.target;
    const checked = (e.target as HTMLInputElement).checked;

    setFormData((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));

    if (validationErrors[name]) {
      setValidationErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  const handleJsonChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setRequestDataJson(value);
    setJsonError('');

    if (value.trim() === '') {
      setFormData((prev) => ({ ...prev, request_data: undefined }));
      return;
    }

    try {
      const parsed = JSON.parse(value);
      setFormData((prev) => ({ ...prev, request_data: parsed }));
    } catch (err) {
      setJsonError('Invalid JSON format');
    }
  };

  const loadTemplate = (template: RequestTemplate) => {
    setFormData((prev) => ({
      ...prev,
      request_type: template.request_type,
      message: template.message,
      request_data: template.request_data,
    }));
    setRequestDataJson(JSON.stringify(template.request_data, null, 2));
    setJsonError('');
  };

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!formData.approver_name?.trim() && !formData.approver_id?.trim()) {
      errors.approver_name = 'Approver name or ID is required';
    }

    if (!formData.requester_name?.trim() && !formData.requester_id?.trim()) {
      errors.requester_name = 'Requester name or ID is required';
    }

    if (!formData.message?.trim()) {
      errors.message = 'Message is required';
    }

    if (formData.use_app_dm && !formData.app_bot_user_id?.trim()) {
      errors.app_bot_user_id = 'App Bot User ID is required when using App DM';
    }

    if (jsonError) {
      errors.request_data = jsonError;
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    await onSubmit(formData);
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
      <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
        Create Approval Request
      </h2>

      <div className="mb-6">
        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          <FileJson className="inline w-4 h-4 mr-1" />
          Quick Templates
        </label>
        <div className="flex flex-wrap gap-2">
          {templates.map((template, index) => (
            <button
              key={index}
              type="button"
              onClick={() => loadTemplate(template)}
              className="px-3 py-1 text-sm bg-slack-purple text-white rounded-md hover:bg-opacity-90 transition-colors"
            >
              {template.name}
            </button>
          ))}
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Bot ID - Optional */}
        <div>
          <label htmlFor="bot_id" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Bot ID (Optional)
          </label>
          <input
            type="text"
            id="bot_id"
            name="bot_id"
            value={formData.bot_id}
            onChange={handleInputChange}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-slack-purple focus:border-transparent dark:bg-gray-700 dark:text-white"
            placeholder="U0AGPDSLH0V"
          />
        </div>

        {/* Approver Name */}
        <div>
          <label htmlFor="approver_name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Approver Name <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="approver_name"
            name="approver_name"
            value={formData.approver_name}
            onChange={handleInputChange}
            className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-slack-purple focus:border-transparent dark:bg-gray-700 dark:text-white ${
              validationErrors.approver_name ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
            }`}
            placeholder="Sarumathi S"
          />
          {validationErrors.approver_name && (
            <p className="mt-1 text-sm text-red-500 flex items-center">
              <AlertCircle className="w-4 h-4 mr-1" />
              {validationErrors.approver_name}
            </p>
          )}
        </div>

        {/* Requester Name */}
        <div>
          <label htmlFor="requester_name" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Requester Name <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="requester_name"
            name="requester_name"
            value={formData.requester_name}
            onChange={handleInputChange}
            className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-slack-purple focus:border-transparent dark:bg-gray-700 dark:text-white ${
              validationErrors.requester_name ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
            }`}
            placeholder="Jeromel Pushparaj"
          />
          {validationErrors.requester_name && (
            <p className="mt-1 text-sm text-red-500 flex items-center">
              <AlertCircle className="w-4 h-4 mr-1" />
              {validationErrors.requester_name}
            </p>
          )}
        </div>

        {/* Request Type */}
        <div>
          <label htmlFor="request_type" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Request Type <span className="text-red-500">*</span>
          </label>
          <select
            id="request_type"
            name="request_type"
            value={formData.request_type}
            onChange={handleInputChange}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-slack-purple focus:border-transparent dark:bg-gray-700 dark:text-white"
          >
            <option value="deployment">Deployment</option>
            <option value="access_request">Access Request</option>
            <option value="code_review">Code Review</option>
            <option value="other">Other</option>
          </select>
        </div>

        {/* Message */}
        <div>
          <label htmlFor="message" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Message <span className="text-red-500">*</span>
          </label>
          <textarea
            id="message"
            name="message"
            value={formData.message}
            onChange={handleInputChange}
            rows={4}
            className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-slack-purple focus:border-transparent dark:bg-gray-700 dark:text-white ${
              validationErrors.message ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
            }`}
            placeholder="Please approve production deployment for api-gateway v2.1.0"
          />
          {validationErrors.message && (
            <p className="mt-1 text-sm text-red-500 flex items-center">
              <AlertCircle className="w-4 h-4 mr-1" />
              {validationErrors.message}
            </p>
          )}
        </div>

        {/* Request Data JSON */}
        <div>
          <label htmlFor="request_data" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            Request Data (JSON, Optional)
          </label>
          <textarea
            id="request_data"
            name="request_data"
            value={requestDataJson}
            onChange={handleJsonChange}
            rows={6}
            className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-slack-purple focus:border-transparent dark:bg-gray-700 dark:text-white font-mono text-sm ${
              jsonError ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
            }`}
            placeholder={`{\n  "service": "api-gateway",\n  "version": "v2.1.0",\n  "environment": "production"\n}`}
          />
          {jsonError && (
            <p className="mt-1 text-sm text-red-500 flex items-center">
              <AlertCircle className="w-4 h-4 mr-1" />
              {jsonError}
            </p>
          )}
        </div>

        {/* Use App DM */}
        <div className="flex items-center">
          <input
            type="checkbox"
            id="use_app_dm"
            name="use_app_dm"
            checked={formData.use_app_dm}
            onChange={handleInputChange}
            className="w-4 h-4 text-slack-purple border-gray-300 rounded focus:ring-slack-purple"
          />
          <label htmlFor="use_app_dm" className="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300">
            Use App DM (Send to approver's DM)
          </label>
        </div>

        {/* App Bot User ID */}
        {formData.use_app_dm && (
          <div>
            <label htmlFor="app_bot_user_id" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              App Bot User ID <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              id="app_bot_user_id"
              name="app_bot_user_id"
              value={formData.app_bot_user_id}
              onChange={handleInputChange}
              className={`w-full px-3 py-2 border rounded-md focus:ring-2 focus:ring-slack-purple focus:border-transparent dark:bg-gray-700 dark:text-white ${
                validationErrors.app_bot_user_id ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'
              }`}
              placeholder="U0AGPDSLH0V"
            />
            {validationErrors.app_bot_user_id && (
              <p className="mt-1 text-sm text-red-500 flex items-center">
                <AlertCircle className="w-4 h-4 mr-1" />
                {validationErrors.app_bot_user_id}
              </p>
            )}
          </div>
        )}

        {/* Submit Button */}
        <button
          type="submit"
          disabled={isLoading}
          className="w-full flex items-center justify-center px-4 py-3 bg-slack-purple text-white font-semibold rounded-md hover:bg-opacity-90 focus:outline-none focus:ring-2 focus:ring-slack-purple focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
        >
          {isLoading ? (
            <>
              <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Submitting...
            </>
          ) : (
            <>
              <Send className="w-5 h-5 mr-2" />
              Submit Approval Request
            </>
          )}
        </button>
      </form>
    </div>
  );
};

export default ApprovalForm;

