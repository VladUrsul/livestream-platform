import api from './api';
import { 
  type AuthResponse, 
  type LoginInput, 
  type RegisterInput,
  type User 
} from '../types/auth.types';

export const authService = {
  register: async (input: RegisterInput): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/register', input);
    return data;
  },

  login: async (input: LoginInput): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/login', input);
    return data;
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout');
  },

  refresh: async (refreshToken: string): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken });
    return data;
  },
  // Validates the stored token and returns fresh user data from the server
  getMe: async (): Promise<User> => {
    const { data } = await api.get<User>('/auth/me');
    return data;
  },
};