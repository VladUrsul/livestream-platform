import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { userService } from '../../services/userService';
import { type SearchResult } from '../../types/user.types';
import { useNotifications } from '../../hooks/useNotifications';
import styles from './Navbar.module.css';

interface NavbarProps {
  onToggleSidebar: () => void;
  sidebarOpen: boolean;
}

export default function Navbar({ onToggleSidebar, sidebarOpen }: NavbarProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [dropdownOpen, setDropdownOpen] = useState(false);

  const [query,       setQuery]       = useState('');
  const [results,     setResults]     = useState<SearchResult[]>([]);
  const [searching,   setSearching]   = useState(false);
  const [searchOpen,  setSearchOpen]  = useState(false);
  const searchRef   = useRef<HTMLDivElement>(null);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { notifications, unreadCount, markAllRead, markRead } = useNotifications();
  const [notifOpen, setNotifOpen] = useState(false);
  const notifRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    if (query.trim().length < 1) {
      setResults([]);
      setSearchOpen(false);
      return;
    }
    debounceRef.current = setTimeout(async () => {
      setSearching(true);
      try {
        const data = await userService.search(query.trim());
        setResults(data);
        setSearchOpen(true);
      } catch {
        setResults([]);
      } finally {
        setSearching(false);
      }
    }, 300);
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current); };
  }, [query]);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) {
        setNotifOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setSearchOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const goToChannel = (username: string) => {
    setQuery('');
    setSearchOpen(false);
    navigate(`/channel/${username}`);
  };

  return (
    <header className={styles.navbar}>

      {/* Left */}
      <div className={styles.left}>
        <button className={styles.menuBtn} onClick={onToggleSidebar} aria-label="Toggle sidebar">
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
      <div className={styles.center} ref={searchRef}>
        <div className={styles.searchWrapper}>
          <span className={styles.searchIcon}>⌕</span>
          <input
            type="text"
            placeholder="Search users..."
            className={styles.searchInput}
            value={query}
            onChange={e => setQuery(e.target.value)}
            onFocus={() => results.length > 0 && setSearchOpen(true)}
          />
          {searching
            ? <span className={styles.searchSpinner} />
            : <span className={styles.searchShortcut}>⌘K</span>
          }
        </div>

        {searchOpen && (
          <div className={styles.searchDropdown}>
            {results.length === 0 ? (
              <div className={styles.searchEmpty}>No users found for "{query}"</div>
            ) : (
              results.map(u => (
                <button
                  key={u.user_id}
                  className={styles.searchResult}
                  onClick={() => goToChannel(u.username)}
                >
                  <div className={styles.searchAvatar}>
                    {u.avatar_url
                      ? <img src={u.avatar_url} alt={u.username} />
                      : u.username[0].toUpperCase()
                    }
                    {u.is_live && <span className={styles.searchLiveDot} />}
                  </div>
                  <div className={styles.searchResultText}>
                    <span className={styles.searchUsername}>@{u.username}</span>
                    {u.display_name && u.display_name !== u.username && (
                      <span className={styles.searchDisplayName}>{u.display_name}</span>
                    )}
                  </div>
                  {u.is_live && <span className={styles.searchLiveBadge}>LIVE</span>}
                </button>
              ))
            )}
          </div>
        )}
      </div>

      {/* Right */}
      <div className={styles.right}>
        <button className={styles.goLiveBtn} onClick={() => navigate('/go-live')}>
          <span className={styles.liveDot} />
          Go Live
        </button>

        <div className={styles.notifMenu} ref={notifRef}>
          <button
            className={styles.iconBtn}
            aria-label="Notifications"
            onClick={() => setNotifOpen(o => !o)}
          >
            <span className={styles.notifIcon}>◎</span>
            {unreadCount > 0 && (
              <span className={styles.notifBadge}>
                {unreadCount > 99 ? '99+' : unreadCount}
              </span>
            )}
          </button>
          
          {notifOpen && (
            <>
              <div className={styles.dropdownOverlay} onClick={() => setNotifOpen(false)} />
              <div className={styles.notifDropdown}>
                <div className={styles.notifHeader}>
                  <span className={styles.notifTitle}>Notifications</span>
                  {unreadCount > 0 && (
                    <button className={styles.markAllBtn} onClick={() => { markAllRead(); }}>
                      Mark all read
                    </button>
                  )}
                </div>
                <div className={styles.notifList}>
                  {notifications.length === 0 ? (
                    <p className={styles.notifEmpty}>No notifications yet</p>
                  ) : (
                    notifications.slice(0, 15).map(n => (
                      <button
                        key={n.id}
                        className={`${styles.notifItem} ${!n.read ? styles.notifItemUnread : ''}`}
                        onClick={() => markRead(n.id)}
                      >
                        <div className={styles.notifItemDot}>
                          {!n.read && <span className={styles.unreadDot} />}
                        </div>
                        <div className={styles.notifItemText}>
                          <p className={styles.notifItemBody}>{n.body}</p>
                          <span className={styles.notifItemTime}>
                            {formatRelative(n.created_at)}
                          </span>
                        </div>
                      </button>
                    ))
                  )}
                </div>
              </div>
            </>
          )}
        </div>

        <div className={styles.userMenu}>
          <button className={styles.avatar} onClick={() => setDropdownOpen(o => !o)}>
            <span className={styles.avatarText}>{user?.username?.[0]?.toUpperCase() ?? '?'}</span>
          </button>
          {dropdownOpen && (
            <>
              <div className={styles.dropdownOverlay} onClick={() => setDropdownOpen(false)} />
              <div className={styles.dropdown}>
                <div className={styles.dropdownHeader}>
                  <span className={styles.dropdownUsername}>@{user?.username}</span>
                  <span className={styles.dropdownEmail}>{user?.email}</span>
                </div>
                <div className={styles.dropdownDivider} />
                <button
                  className={styles.dropdownItem}
                  onClick={() => { setDropdownOpen(false); navigate(`/channel/${user?.username}`); }}
                >
                  My Channel
                </button>
                <Link to="/go-live" className={styles.dropdownItem} onClick={() => setDropdownOpen(false)}>
                  Go Live
                </Link>
                <div className={styles.dropdownDivider} />
                <button className={styles.dropdownItemDanger} onClick={logout}>Sign out</button>
              </div>
            </>
          )}
        </div>
      </div>
    </header>
  );
}

function formatRelative(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins  = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days  = Math.floor(diff / 86400000);
  if (mins  < 1)  return 'just now';
  if (mins  < 60) return `${mins}m ago`;
  if (hours < 24) return `${hours}h ago`;
  return `${days}d ago`;
}