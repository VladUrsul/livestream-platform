import { NavLink } from 'react-router-dom';
import styles from './Sidebar.module.css';

interface SidebarProps {
  open: boolean;
}

const navItems = [
  { icon: '⊞', label: 'Dashboard',      to: '/dashboard' },
  { icon: '◉', label: 'Live Now',        to: '/live',      badge: '12' },
  { icon: '▶', label: 'Browse',          to: '/browse' },
  { icon: '◈', label: 'Following',       to: '/following' },
  { icon: '♡', label: 'Subscriptions',   to: '/subscriptions' },
];

const secondaryItems = [
  { icon: '◎', label: 'My Channel',    to: '/channel' },
  { icon: '⊙', label: 'Analytics',     to: '/analytics' },
  { icon: '◇', label: 'Settings',      to: '/settings' },
];

// Hardcoded mock live streams for now.
// Will be replaced with real data from stream-service later.
const liveStreams = [
  { id: 1, username: 'techwave',    title: 'Building a Rust compiler', viewers: '2.4k', category: 'Programming' },
  { id: 2, username: 'pixelcraft',  title: 'Pixel art speed drawing',  viewers: '891',  category: 'Art' },
  { id: 3, username: 'synthwave99', title: 'Lo-fi beats live session', viewers: '5.1k', category: 'Music' },
  { id: 4, username: 'cloudnative', title: 'K8s deep dive — Day 3',    viewers: '1.2k', category: 'DevOps' },
];

export default function Sidebar({ open }: SidebarProps) {
  return (
    <>
      {/* Backdrop on mobile */}
      <div className={`${styles.backdrop} ${open ? styles.backdropVisible : ''}`} />

      <aside className={`${styles.sidebar} ${open ? styles.sidebarOpen : styles.sidebarClosed}`}>
        <div className={styles.inner}>

          {/* Main nav */}
          <nav className={styles.nav}>
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  `${styles.navItem} ${isActive ? styles.navItemActive : ''}`
                }
              >
                <span className={styles.navIcon}>{item.icon}</span>
                <span className={styles.navLabel}>{item.label}</span>
                {item.badge && (
                  <span className={styles.navBadge}>{item.badge}</span>
                )}
              </NavLink>
            ))}
          </nav>

          <div className={styles.divider} />

          {/* Live streams section */}
          <div className={styles.section}>
            <div className={styles.sectionHeader}>
              <span className={styles.sectionTitle}>LIVE NOW</span>
              <span className={styles.liveIndicator}>
                <span className={styles.liveDot} />
                {liveStreams.length}
              </span>
            </div>

            <div className={styles.streamList}>
              {liveStreams.map((stream) => (
                <a key={stream.id} href={`/stream/${stream.username}`} className={styles.streamItem}>
                  <div className={styles.streamAvatar}>
                    {stream.username[0].toUpperCase()}
                    <span className={styles.streamLiveDot} />
                  </div>
                  <div className={styles.streamInfo}>
                    <span className={styles.streamUsername}>@{stream.username}</span>
                    <span className={styles.streamTitle}>{stream.title}</span>
                    <div className={styles.streamMeta}>
                      <span className={styles.streamViewers}>◎ {stream.viewers}</span>
                      <span className={styles.streamCategory}>{stream.category}</span>
                    </div>
                  </div>
                </a>
              ))}
            </div>
          </div>

          <div className={styles.divider} />

          {/* Secondary nav */}
          <nav className={styles.nav}>
            {secondaryItems.map((item) => (
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