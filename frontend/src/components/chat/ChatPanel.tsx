import { useEffect, useRef, useState } from 'react';
import { 
    ChatSocket, 
    type ChatMessage 
} from '../../services/chatService';
import { useAuth } from '../../hooks/useAuth';
import styles from './ChatPanel.module.css';

interface ChatPanelProps {
  roomID: string;
  isOwner?: boolean;
}

export default function ChatPanel({ roomID, isOwner = false }: ChatPanelProps) {
  const { accessToken } = useAuth();
  const socketRef  = useRef<ChatSocket | null>(null);
  const bottomRef  = useRef<HTMLDivElement>(null);

  const [messages,  setMessages]  = useState<ChatMessage[]>([]);
  const [input,     setInput]     = useState('');
  const [slowMode,  setSlowMode]  = useState(0);
  const [cooldown,  setCooldown]  = useState(0);
  const [connected, setConnected] = useState(false);
  const [error,     setError]     = useState<string | null>(null);
  const cooldownRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // ── WebSocket ─────────────────────────────────────────────────────
  useEffect(() => {
    if (!accessToken) return;

    const socket = new ChatSocket(roomID, accessToken);
    socketRef.current = socket;

    const unsub = socket.on((event) => {
      if (event.type === 'history') {
        setMessages(event.history ?? []);
        setConnected(true);
      } else if (event.type === 'message') {
        setMessages(prev => [...prev, event.message]);
      } else if (event.type === 'slow_mode') {
        setSlowMode(event.slow_mode);
      } else if (event.type === 'error') {
        setError(event.error);
        setTimeout(() => setError(null), 3000);
      }
    });

    socket.connect();

    return () => {
      unsub();
      socket.disconnect();
    };
  }, [roomID, accessToken]);

  // Auto-scroll to bottom on new messages
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // ── Send ──────────────────────────────────────────────────────────
  const send = () => {
    const content = input.trim();
    if (!content || cooldown > 0) return;
    socketRef.current?.send(content);
    setInput('');

    // Start local cooldown
    if (slowMode > 0) {
      setCooldown(slowMode);
      if (cooldownRef.current) clearInterval(cooldownRef.current);
      cooldownRef.current = setInterval(() => {
        setCooldown(prev => {
          if (prev <= 1) {
            clearInterval(cooldownRef.current!);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  };

  // ── Slow mode setter (owner only) ─────────────────────────────────
  const setSlowModeRemote = async (seconds: number) => {
    try {
      const token = accessToken;
      await fetch(`http://localhost:8080/api/v1/chat/${roomID}/slow-mode`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ seconds }),
      });
    } catch {}
  };

  return (
    <div className={styles.panel}>

      {/* Header */}
      <div className={styles.header}>
        <span className={styles.headerTitle}>Chat</span>
        {slowMode > 0 && (
          <span className={styles.slowBadge}>⏱ {slowMode}s</span>
        )}
        {isOwner && (
          <div className={styles.slowControls}>
            {[0, 5, 10, 30].map(s => (
              <button
                key={s}
                className={`${styles.slowBtn} ${slowMode === s ? styles.slowBtnActive : ''}`}
                onClick={() => setSlowModeRemote(s)}
              >
                {s === 0 ? 'Off' : `${s}s`}
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Messages */}
      <div className={styles.messages}>
        {!connected && (
          <div className={styles.connecting}>
            <span className={styles.connectingSpinner} />
            Connecting...
          </div>
        )}
        {connected && messages.length === 0 && (
          <p className={styles.emptyChat}>No messages yet. Say hi!</p>
        )}
        {messages.map(msg => (
          <div key={msg.id} className={styles.message}>
            <span className={styles.msgUsername}>@{msg.username}</span>
            <span className={styles.msgContent}>{msg.content}</span>
          </div>
        ))}
        {error && (
          <div className={styles.errorToast}>{error}</div>
        )}
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className={styles.inputArea}>
        <input
          className={styles.input}
          placeholder={cooldown > 0 ? `Wait ${cooldown}s...` : 'Send a message...'}
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={cooldown > 0}
          maxLength={500}
        />
        <button
          className={styles.sendBtn}
          onClick={send}
          disabled={!input.trim() || cooldown > 0}
        >
          {cooldown > 0 ? cooldown : '↑'}
        </button>
      </div>

    </div>
  );
}