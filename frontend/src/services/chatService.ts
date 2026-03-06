const WS_BASE = 'ws://localhost:8080';

export interface ChatMessage {
  id: string;
  room_id: string;
  user_id: string;
  username: string;
  content: string;
  created_at: string;
}

export type WSEvent =
  | { type: 'message';   message: ChatMessage }
  | { type: 'history';   history: ChatMessage[] | null }
  | { type: 'error';     error: string }
  | { type: 'slow_mode'; slow_mode: number };

export class ChatSocket {
  private ws: WebSocket | null = null;
  private roomID: string;
  private token: string;
  private listeners: Array<(e: WSEvent) => void> = [];
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private closed = false;

  constructor(roomID: string, token: string) {
    this.roomID = roomID;
    this.token  = token;
  }

  connect() {
    this.closed = false;
    const url = `${WS_BASE}/ws/${this.roomID}?token=${encodeURIComponent(this.token)}`;
    this.ws = new WebSocket(url);

    this.ws.onmessage = (e) => {
      try {
        const event: WSEvent = JSON.parse(e.data);
        // copy array before iterating so unsub mid-dispatch is safe
        [...this.listeners].forEach(l => l(event));
      } catch {}
    };

    this.ws.onclose = () => {
      if (!this.closed) {
        this.reconnectTimer = setTimeout(() => this.connect(), 3000);
      }
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  send(content: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(content);
    }
  }

  on(listener: (e: WSEvent) => void) {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener);
    };
  }

  disconnect() {
    this.closed = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.ws?.close();
    this.listeners = [];
  }
}