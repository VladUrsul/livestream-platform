import { useEffect, useRef, useState } from 'react';
import { useAuth } from './useAuth';
import {
  NotificationSocket,
  notificationApi,
  type Notification,
} from '../services/notificationService';

export const useNotifications = () => {
  const { accessToken, isAuthenticated } = useAuth();
  const socketRef = useRef<NotificationSocket | null>(null);

  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount,   setUnreadCount]   = useState(0);

  // Connect WebSocket when authenticated
  useEffect(() => {
    if (!isAuthenticated || !accessToken) return;

    const socket = new NotificationSocket(accessToken);
    socketRef.current = socket;

    const unsub = socket.on((event) => {
      if (event.type === 'unread_count') {
        setUnreadCount(event.unread_count);
      } else if (event.type === 'notification') {
        setUnreadCount(event.unread_count);
        setNotifications(prev => [event.notification, ...prev]);
      }
    });

    socket.connect();

    // Load initial notifications
    notificationApi.getAll().then(data => {
      setNotifications(data.notifications ?? []);
      setUnreadCount(data.unread_count);
    }).catch(() => {});

    return () => {
      unsub();
      socket.disconnect();
    };
  }, [isAuthenticated, accessToken]);

  const markAllRead = async () => {
    await notificationApi.markAllRead();
    setUnreadCount(0);
    setNotifications(prev => prev.map(n => ({ ...n, read: true })));
  };

  const markRead = async (id: string) => {
    await notificationApi.markRead(id);
    setNotifications(prev =>
      prev.map(n => n.id === id ? { ...n, read: true } : n)
    );
    setUnreadCount(prev => Math.max(0, prev - 1));
  };

  return { notifications, unreadCount, markAllRead, markRead };
};