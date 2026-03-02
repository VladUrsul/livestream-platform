import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import styles from './Navbar.module.css';

interface NavbarProps {
  onToggleSidebar: () => void;
  sidebarOpen: boolean;
}

export default function Navbar({ onToggleSidebar, sidebarOpen }: NavbarProps) {
  const { user, logout } = useAuth();
  const [dropdownOpen, setDropdownOpen] = useState(false);

  return (
    <header className={styles.navbar}>
      {/* Left — logo + sidebar toggle */}
      <div className={styles.left}>
        <button
          className={styles.menuBtn}
          onClick={onToggleSidebar}
          aria-label="Toggle sidebar"
        >
          <span className={`${styles.menuIcon} ${sidebarOpen ? styles.menuIconOpen : ''}`}>
            <span /><span /><span />
          </span>
        </button>

        <Link to="/dashboard" className={styles.logo}>
          <span className={styles.logoSymbol}>◈</span>
          <span className={styles.logoText}>STREAMR</span>
        </Link>
      </div>

      {/* Center — search */}
      <div className={styles.center}>
        <div className={styles.searchWrapper}>
          <span className={styles.searchIcon}>⌕</span>
          <input
            type="text"
            placeholder="Search streams, channels..."
            className={styles.searchInput}
          />
          <span className={styles.searchShortcut}>⌘K</span>
        </div>
      </div>

      {/* Right — actions + user */}
      <div className={styles.right}>
        <button className={styles.goLiveBtn}>
          <span className={styles.liveDot} />
          Go Live
        </button>

        <button className={styles.iconBtn} aria-label="Notifications">
          <span className={styles.notifIcon}>◎</span>
          <span className={styles.notifBadge}>3</span>
        </button>

        <div className={styles.userMenu}>
          <button
            className={styles.avatar}
            onClick={() => setDropdownOpen(!dropdownOpen)}
          >
            <span className={styles.avatarText}>
              {user?.username?.[0]?.toUpperCase() ?? '?'}
            </span>
          </button>

          {dropdownOpen && (
            <>
              <div
                className={styles.dropdownOverlay}
                onClick={() => setDropdownOpen(false)}
              />
              <div className={styles.dropdown}>
                <div className={styles.dropdownHeader}>
                  <span className={styles.dropdownUsername}>@{user?.username}</span>
                  <span className={styles.dropdownEmail}>{user?.email}</span>
                </div>
                <div className={styles.dropdownDivider} />
                <Link to="/settings" className={styles.dropdownItem} onClick={() => setDropdownOpen(false)}>
                  Settings
                </Link>
                <Link to="/channel" className={styles.dropdownItem} onClick={() => setDropdownOpen(false)}>
                  My Channel
                </Link>
                <div className={styles.dropdownDivider} />
                <button className={styles.dropdownItemDanger} onClick={logout}>
                  Sign out
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </header>
  );
}