import api from './api';

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  actor_id: string;
  actor_name: string;
  read: boolean;
  created_at: string;
}

export type NotificationWSEvent =
  | { type: 'notification'; notification: Notification; unread_count: number }
  | { type: 'unread_count'; unread_count: number };

export class NotificationSocket {
  private ws: WebSocket | null = null;
  private token: string;
  private listeners: Array<(e: NotificationWSEvent) => void> = [];
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private closed = false;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;

  constructor(token: string) {
    this.token = token;
  }

  connect() {
    this.closed = false;
    const url = `ws://localhost:8080/ws/notifications?token=${encodeURIComponent(this.token)}`;
    this.ws = new WebSocket(url);

    this.ws.onmessage = (e) => {
      try {
        const event: NotificationWSEvent = JSON.parse(e.data);
        [...this.listeners].forEach(l => l(event));
      } catch {}
    };

    this.ws.onopen = () => {
      this.reconnectAttempts = 0; // reset on successful connect
    };

    this.ws.onclose = (e) => {
      if (!this.closed && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        this.reconnectTimer = setTimeout(() => this.connect(), 3000);
      }
    };

    this.ws.onerror = () => this.ws?.close();
  }

  on(listener: (e: NotificationWSEvent) => void) {
    this.listeners.push(listener);
    return () => { this.listeners = this.listeners.filter(l => l !== listener); };
  }

  disconnect() {
    this.closed = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.ws?.close();
    this.listeners = [];
  }
}

export const notificationApi = {
  getAll: async (limit = 20, offset = 0) => {
    const { data } = await api.get<{
      notifications: Notification[];
      unread_count: number;
    }>(`/notifications?limit=${limit}&offset=${offset}`);
    return data;
  },

  markAllRead: async () => {
    await api.put('/notifications/read-all');
  },

  markRead: async (id: string) => {
    await api.put(`/notifications/${id}/read`);
  },
};