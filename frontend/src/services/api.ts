import axios, { AxiosError } from 'axios';
import { GenericApprovalRequest, ApprovalResponse } from '../types/approval';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8083';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
});

export const approvalApi = {
  createApprovalRequest: async (
    request: GenericApprovalRequest
  ): Promise<ApprovalResponse> => {
    try {
      const response = await apiClient.post<ApprovalResponse>(
        '/api/v1/approval/generic',
        request
      );
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError<ApprovalResponse>;
        if (axiosError.response?.data) {
          return axiosError.response.data;
        }
        return {
          success: false,
          error: axiosError.message || 'Network error occurred',
        };
      }
      return {
        success: false,
        error: 'An unexpected error occurred',
      };
    }
  },

  healthCheck: async (): Promise<boolean> => {
    try {
      const response = await apiClient.get('/health');
      return response.status === 200;
    } catch {
      return false;
    }
  },
};

export default apiClient;

