import axios from 'axios';
import type { AxiosError, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import {
  clearStoredTokens,
  getAccessToken,
  getRefreshToken,
  setStoredTokens,
  type SessionTokens,
} from '../store/session';

export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
});

const refreshClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
});

type AuthLifecycleHandlers = {
  onTokensUpdated?: (tokens: SessionTokens) => void;
  onAuthFailure?: () => void;
};

type RetriableConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
};

let lifecycleHandlers: AuthLifecycleHandlers = {};
let refreshRequest: Promise<SessionTokens> | null = null;

apiClient.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiEnvelope<unknown>>) => {
    const originalRequest = error.config as RetriableConfig | undefined;
    const status = error.response?.status;

    if (!originalRequest || status !== 401 || originalRequest._retry) {
      return Promise.reject(error);
    }

    if (originalRequest.url?.includes('/auth/login') || originalRequest.url?.includes('/auth/register')) {
      return Promise.reject(error);
    }

    const refreshToken = getRefreshToken();
    if (!refreshToken) {
      clearStoredTokens();
      lifecycleHandlers.onAuthFailure?.();
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      const refreshedTokens = await requestTokenRefresh(refreshToken);
      originalRequest.headers = originalRequest.headers ?? {};
      originalRequest.headers.Authorization = `Bearer ${refreshedTokens.access_token}`;
      return apiClient(originalRequest);
    } catch (refreshError) {
      clearStoredTokens();
      lifecycleHandlers.onAuthFailure?.();
      return Promise.reject(refreshError);
    }
  },
);

export async function unwrapData<T>(
  request: Promise<AxiosResponse<ApiEnvelope<T>>>,
): Promise<T> {
  const response = await request;
  return response.data.data;
}

export async function requestTokenRefresh(refreshToken: string): Promise<SessionTokens> {
  if (!refreshRequest) {
    refreshRequest = refreshClient
      .post<ApiEnvelope<SessionTokens>>('/auth/refresh', {
        refresh_token: refreshToken,
      })
      .then((response) => {
        const tokens = response.data.data;
        setStoredTokens(tokens);
        lifecycleHandlers.onTokensUpdated?.(tokens);
        return tokens;
      })
      .finally(() => {
        refreshRequest = null;
      });
  }

  return refreshRequest;
}

export function registerAuthLifecycleHandlers(handlers: AuthLifecycleHandlers) {
  lifecycleHandlers = handlers;
}

export function withAuthHeader(config: AxiosRequestConfig = {}): AxiosRequestConfig {
  const token = getAccessToken();
  return {
    ...config,
    headers: {
      ...config.headers,
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  };
}

export default apiClient;
