import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { streamService } from '../services/streamService';
import { type StreamInfo } from '../types/stream.types';
import styles from './Dashboard.module.css';

const categories = [
  { name: 'Programming', icon: '⟨⟩' },
  { name: 'Gaming',      icon: '◈'  },
  { name: 'Music',       icon: '♩'  },
  { name: 'Art',         icon: '◎'  },
  { name: 'DevOps',      icon: '⊙'  },
  { name: 'Design',      icon: '◇'  },
];

const categoryColors: Record<string, string> = {
  Programming: '#3b82f6',
  Art:         '#a855f7',
  Music:       '#ec4899',
  DevOps:      '#22c55e',
  GameDev:     '#f97316',
  Design:      '#e8ff47',
  General:     '#6b7280',
};

export default function Dashboard() {
  const { user } = useAuth();
  const navigate  = useNavigate();

  const [liveStreams, setLiveStreams] = useState<StreamInfo[]>([]);
  const [loading,    setLoading]     = useState(true);

  const hour     = new Date().getHours();
  const greeting = hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening';

  useEffect(() => {
    const load = async () => {
      try {
        const streams = await streamService.getLiveStreams();
        setLiveStreams(streams ?? []);
      } catch {
        setLiveStreams([]);
      } finally {
        setLoading(false);
      }
    };
    load();
    const interval = setInterval(load, 30_000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className={styles.page}>

      {/* ── Hero ── */}
      <div className={styles.hero}>
        <div className={styles.heroText}>
          <p className={styles.heroGreeting}>{greeting}</p>
          <h1 className={styles.heroTitle}>
            Welcome back,{' '}
            <span className={styles.heroUsername}>@{user?.username}</span>
          </h1>
          <p className={styles.heroSub}>
            <span className={styles.liveCount}>◉ {liveStreams.length} streams</span> live right now
          </p>
        </div>

        <div className={styles.heroStats}>
          <div className={styles.heroStat}>
            <span className={styles.heroStatNum}>0</span>
            <span className={styles.heroStatLabel}>Followers</span>
          </div>
          <div className={styles.heroStatDivider} />
          <div className={styles.heroStat}>
            <span className={styles.heroStatNum}>0</span>
            <span className={styles.heroStatLabel}>Following</span>
          </div>
          <div className={styles.heroStatDivider} />
          <div className={styles.heroStat}>
            <span className={styles.heroStatNum}>0</span>
            <span className={styles.heroStatLabel}>Hours watched</span>
          </div>
        </div>
      </div>

      {/* ── Categories ── */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Browse by category</h2>
        </div>
        <div className={styles.categoryGrid}>
          {categories.map(cat => (
            <div key={cat.name} className={styles.categoryCard}>
              <span className={styles.categoryIcon}>{cat.icon}</span>
              <span className={styles.categoryName}>{cat.name}</span>
            </div>
          ))}
        </div>
      </section>

      {/* ── Live streams ── */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>
            Live now
            <span className={styles.liveBadge}>
              <span className={styles.liveDot} />
              LIVE
            </span>
          </h2>
          <span className={styles.sectionCount}>
            {loading ? '...' : `${liveStreams.length} online`}
          </span>
        </div>

        {loading ? (
          <div className={styles.streamGrid}>
            {[...Array(4)].map((_, i) => (
              <div key={i} className={styles.skeletonCard} />
            ))}
          </div>
        ) : liveStreams.length === 0 ? (
          <div className={styles.emptyState}>
            <span className={styles.emptyIcon}>◎</span>
            <p>No one is live right now.</p>
            <button className={styles.goLivePrompt} onClick={() => navigate('/go-live')}>
              Be the first — Go Live
            </button>
          </div>
        ) : (
          <div className={styles.streamGrid}>
            {liveStreams.map(stream => (
              <button
                key={stream.id}
                className={styles.streamCard}
                onClick={() => navigate(`/channel/${stream.username}`)}
              >
                <div className={styles.streamThumbnail}>
                  <div
                    className={styles.thumbnailBg}
                    style={{
                      background: `linear-gradient(135deg, ${categoryColors[stream.category] ?? '#374151'}22, #111)`,
                    }}
                  />
                  <div className={styles.thumbnailOverlay}>
                    <span className={styles.thumbnailIcon}>▶</span>
                  </div>
                  <div className={styles.streamBadges}>
                    <span className={styles.livePill}>
                      <span className={styles.livePillDot} />
                      LIVE
                    </span>
                    <span className={styles.viewersPill}>
                      ◎ {stream.viewer_count.toLocaleString()}
                    </span>
                  </div>
                </div>

                <div className={styles.streamCardInfo}>
                  <div className={styles.streamCardAvatar}>
                    {stream.username[0].toUpperCase()}
                  </div>
                  <div className={styles.streamCardText}>
                    <p className={styles.streamCardTitle}>{stream.title}</p>
                    <p className={styles.streamCardUsername}>@{stream.username}</p>
                    <span
                      className={styles.streamCardCategory}
                      style={{
                        borderColor: `${categoryColors[stream.category] ?? '#374151'}44`,
                        color: categoryColors[stream.category] ?? '#6b7280',
                      }}
                    >
                      {stream.category}
                    </span>
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </section>

    </div>
  );
}