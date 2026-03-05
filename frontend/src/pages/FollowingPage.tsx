import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { userService } from '../services/userService';
import { type SearchResult } from '../types/user.types';
import styles from './FollowingPage.module.css';

export default function FollowingPage() {
  const navigate = useNavigate();
  const [users,   setUsers]   = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    userService.getFollowing()
      .then(setUsers)
      .catch(() => setUsers([]))
      .finally(() => setLoading(false));
  }, []);

  const handleUnfollow = async (username: string) => {
    try {
      await userService.unfollow(username);
      setUsers(prev => prev.filter(u => u.username !== username));
    } catch {}
  };

  return (
    <div className={styles.page}>

      <div className={styles.header}>
        <h1 className={styles.title}>Following</h1>
        <span className={styles.count}>
          {loading ? '...' : `${users.length} accounts`}
        </span>
      </div>

      {loading ? (
        <div className={styles.grid}>
          {[...Array(6)].map((_, i) => (
            <div key={i} className={styles.skeletonCard} />
          ))}
        </div>
      ) : users.length === 0 ? (
        <div className={styles.emptyState}>
          <span className={styles.emptyIcon}>◈</span>
          <p className={styles.emptyTitle}>Not following anyone yet</p>
          <p className={styles.emptySub}>Find streamers to follow from the dashboard or search.</p>
          <button className={styles.browseBtn} onClick={() => navigate('/dashboard')}>
            Browse streams
          </button>
        </div>
      ) : (
        <div className={styles.grid}>
          {users.map(u => (
            <div key={u.user_id} className={styles.card}>
              <button
                className={styles.cardMain}
                onClick={() => navigate(`/channel/${u.username}`)}
              >
                <div className={styles.cardAvatar}>
                  {u.avatar_url
                    ? <img src={u.avatar_url} alt={u.username} />
                    : u.username[0].toUpperCase()
                  }
                  {u.is_live && <span className={styles.liveDot} />}
                </div>

                <div className={styles.cardInfo}>
                  <div className={styles.cardNameRow}>
                    <span className={styles.cardDisplayName}>
                      {u.display_name || u.username}
                    </span>
                    {u.is_live && <span className={styles.liveBadge}>LIVE</span>}
                  </div>
                  <span className={styles.cardUsername}>@{u.username}</span>
                  <span className={styles.cardFollowers}>
                    {u.followers.toLocaleString()} followers
                  </span>
                </div>
              </button>

              <div className={styles.cardActions}>
                {u.is_live && (
                  <button
                    className={styles.watchBtn}
                    onClick={() => navigate(`/channel/${u.username}`)}
                  >
                    <span className={styles.watchDot} />
                    Watch
                  </button>
                )}
                <button
                  className={styles.unfollowBtn}
                  onClick={() => handleUnfollow(u.username)}
                >
                  Unfollow
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}