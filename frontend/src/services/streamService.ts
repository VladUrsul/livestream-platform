import api from './api';
import { 
    type StreamInfo, 
    type StreamKeyResponse, 
    type StreamSettings 
} from '../types/stream.types';

export const streamService = {
  getStreamKey: async (): Promise<StreamKeyResponse> => {
    const { data } = await api.get<StreamKeyResponse>('/streams/key');
    return data;
  },

  rotateStreamKey: async (): Promise<StreamKeyResponse> => {
    const { data } = await api.post<StreamKeyResponse>('/streams/key/rotate');
    return data;
  },

  updateSettings: async (settings: StreamSettings): Promise<void> => {
    await api.put('/streams/settings', settings);
  },

  getStreamInfo: async (username: string): Promise<StreamInfo> => {
    const { data } = await api.get<StreamInfo>(`/streams/${username}`);
    return data;
  },

  getLiveStreams: async (): Promise<StreamInfo[]> => {
    const { data } = await api.get<{ streams: StreamInfo[] }>('/streams/live');
    return data.streams;
  },

  joinStream: async (username: string): Promise<void> => {
    await api.post(`/streams/${username}/join`);
  },

  leaveStream: async (username: string): Promise<void> => {
    await api.post(`/streams/${username}/leave`);
  },
};