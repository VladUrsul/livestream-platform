import { useEffect, useState } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { streamService } from '../../services/streamService';
import { type StreamInfo } from '../../types/stream.types';
import styles from './Sidebar.module.css';

interface SidebarProps { open: boolean; }

const navItems = [
  { icon: '⊞', label: 'Dashboard',    to: '/dashboard' },
  { icon: '▶', label: 'Browse',        to: '/browse' },
  { icon: '◈', label: 'Following',     to: '/following' },
  { icon: '♡', label: 'Subscriptions', to: '/subscriptions' },
];

const secondaryItems = [
  { icon: '⊙', label: 'Analytics', to: '/analytics' },
  { icon: '◇', label: 'Settings',  to: '/settings' },
];

export default function Sidebar({ open }: SidebarProps) {
  const { user } = useAuth();
  const navigate  = useNavigate();
  const [liveStreams, setLiveStreams] = useState<StreamInfo[]>([]);

  useEffect(() => {
    const load = async () => {
      try {
        const streams = await streamService.getLiveStreams();
        setLiveStreams(streams ?? []);
      } catch {
        setLiveStreams([]);
      }
    };
    load();
    const interval = setInterval(load, 30_000);
    return () => clearInterval(interval);
  }, []);

  return (
    <>
      <div className={`${styles.backdrop} ${open ? styles.backdropVisible : ''}`} />

      <aside className={`${styles.sidebar} ${open ? styles.sidebarOpen : styles.sidebarClosed}`}>
        <div className={styles.inner}>

          {/* Main nav */}
          <nav className={styles.nav}>
            {navItems.map(item => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `${styles.navItem} ${isActive ? styles.navItemActive : ''}`
                }
              >
                <span className={styles.navIcon}>{item.icon}</span>
                <span className={styles.navLabel}>{item.label}</span>
              </NavLink>
            ))}
          </nav>

          <div className={styles.divider} />

          {/* Live streams */}
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <span className={styles.sectionTitle}>LIVE NOW</span>
              <span className={styles.liveIndicator}>
                <span className={styles.liveDot} />
                {liveStreams.length}
              </span>
            </div>

            <div className={styles.streamList}>
              {liveStreams.length === 0 ? (
                <p className={styles.emptyLive}>No streams online</p>
              ) : (
                liveStreams.map(stream => (
                  <button
                    key={stream.id}
                    className={styles.streamItem}
                    onClick={() => navigate(`/channel/${stream.username}`)}
                  >
                    <div className={styles.streamAvatar}>
                      {stream.username[0].toUpperCase()}
                      <span className={styles.streamLiveDot} />
                    </div>
                    <div className={styles.streamInfo}>
                      <span className={styles.streamUsername}>@{stream.username}</span>
                      <span className={styles.streamTitle}>{stream.title}</span>
                      <div className={styles.streamMeta}>
                        <span className={styles.streamViewers}>◎ {stream.viewer_count.toLocaleString()}</span>
                        <span className={styles.streamCategory}>{stream.category}</span>
                      </div>
                    </div>
                  </button>
                ))
              )}
            </div>
          </div>

          <div className={styles.divider} />

          {/* Secondary nav */}
          <nav className={styles.nav}>
            <button
              className={styles.navItem}
              onClick={() => navigate(`/channel/${user?.username}`)}
            >
              <span className={styles.navIcon}>◎</span>
              <span className={styles.navLabel}>My Channel</span>
            </button>
            {secondaryItems.map(item => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `${styles.navItem} ${isActive ? styles.navItemActive : ''}`
                }
              >
                <span className={styles.navIcon}>{item.icon}</span>
                <span className={styles.navLabel}>{item.label}</span>
              </NavLink>
            ))}
          </nav>

        </div>
      </aside>
    </>
  );
}