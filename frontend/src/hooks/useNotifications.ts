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

  useEffect(() => {
    if (!isAuthenticated || !accessToken) {
      // Clean up if logged out
      if (socketRef.current) {
        socketRef.current.disconnect();
        socketRef.current = null;
      }
      return;
    }

    // Disconnect previous socket (e.g. token rotated)
    if (socketRef.current) {
      socketRef.current.disconnect();
      socketRef.current = null;
    }

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

    notificationApi.getAll().then(data => {
      setNotifications(data.notifications ?? []);
      setUnreadCount(data.unread_count);
    }).catch(() => {});

    return () => {
      unsub();
      socket.disconnect();
      socketRef.current = null;
    };
  }, [isAuthenticated, accessToken]); // ← re-runs when token refreshes

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