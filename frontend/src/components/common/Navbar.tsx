import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { userService } from '../../services/userService';
import { type SearchResult } from '../../types/user.types';
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

        <button className={styles.iconBtn} aria-label="Notifications">
          <span className={styles.notifIcon}>◎</span>
          <span className={styles.notifBadge}>3</span>
        </button>

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