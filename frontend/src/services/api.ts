import axios, {
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from 'axios';
import { store } from '../store';
import { setTokens, logout } from '../store/authSlice';

const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

const api: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const { accessToken } = store.getState().auth;
  if (accessToken && config.headers) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      const { refreshToken } = store.getState().auth;
      if (!refreshToken) {
        store.dispatch(logout());
        return Promise.reject(error);
      }

      try {
        const { data } = await axios.post(`${BASE_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        store.dispatch(setTokens({
          accessToken:  data.access_token,
          refreshToken: data.refresh_token,
        }));

        originalRequest.headers.Authorization = `Bearer ${data.access_token}`;
        return api(originalRequest);
      } catch {
        store.dispatch(logout());
        return Promise.reject(error);
      }
    }

    return Promise.reject(error);
  }
);

export default api;