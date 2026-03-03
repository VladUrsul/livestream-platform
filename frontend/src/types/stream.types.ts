export type StreamStatus = 'offline' | 'live' | 'ended';

export interface StreamInfo {
  id: string;
  user_id: string;
  username: string;
  title: string;
  category: string;
  status: StreamStatus;
  viewer_count: number;
  started_at?: string;
  hls_url?: string;
}

export interface StreamKeyResponse {
  stream_key: string;
  rtmp_url: string;
  obs_url: string;
}

export interface StreamSettings {
  title: string;
  category: string;
}