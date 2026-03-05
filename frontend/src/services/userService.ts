import api from './api';
import { 
    type Profile, 
    type SearchResult 
} from '../types/user.types';

export const userService = {
  search: async (query: string): Promise<SearchResult[]> => {
    const { data } = await api.get<{ users: SearchResult[] }>(`/users/search?q=${encodeURIComponent(query)}`);
    return data.users ?? [];
  },

  getProfile: async (username: string): Promise<Profile> => {
    const { data } = await api.get<Profile>(`/users/${username}`);
    return data;
  },

  getMe: async (): Promise<Profile> => {
    const { data } = await api.get<Profile>('/users/me');
    return data;
  },

  updateProfile: async (input: { display_name?: string; bio?: string; avatar_url?: string }): Promise<Profile> => {
    const { data } = await api.put<Profile>('/users/me', input);
    return data;
  },

  follow: async (username: string): Promise<void> => {
    await api.post(`/users/${username}/follow`);
  },

  unfollow: async (username: string): Promise<void> => {
    await api.delete(`/users/${username}/follow`);
  },

  getFollowing: async (): Promise<SearchResult[]> => {
    const { data } = await api.get<{ users: SearchResult[] }>('/users/me/following');
    return data.users ?? [];
  },

  isFollowing: async (username: string): Promise<boolean> => {
  try {
    const { data } = await api.get<{ following: boolean }>(`/users/${username}/follow`);
    return data.following;
  } catch {
    return false;
  }
},
};