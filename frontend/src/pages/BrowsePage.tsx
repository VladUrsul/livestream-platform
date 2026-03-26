import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { streamService } from '../services/streamService';
import { type StreamInfo } from '../types/stream.types';
import styles from './BrowsePage.module.css';

const CATEGORY_PILLS: Array<{ name: string; icon: string }> = [
  { name: 'All', icon: '⊞' },
  { name: 'Programming', icon: '⟨⟩' },
  { name: 'Gaming', icon: '◈' },
  { name: 'Music', icon: '♩' },
  { name: 'Art', icon: '◎' },
  { name: 'DevOps', icon: '⊙' },
  { name: 'Design', icon: '◇' },
  { name: 'General', icon: '◎' },
];

const normalize = (s: string) => s.toLowerCase().trim();

export default function BrowsePage() {
  const navigate = useNavigate();

  const [streams, setStreams] = useState<StreamInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [activeCategory, setActiveCategory] = useState('All');

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      try {
        const live = await streamService.getLiveStreams();
        if (cancelled) return;
        setStreams(live ?? []);
      } catch {
        if (cancelled) return;
        setStreams([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    load();
    const interval = setInterval(load, 30_000);
    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, []);

  const filtered = useMemo(() => {
    const q = normalize(query);
    return streams.filter((s) => {
      if (activeCategory !== 'All' && s.category !== activeCategory) return false;
      if (!q) return true;
      return (
        normalize(s.title).includes(q) ||
        normalize(s.username).includes(q) ||
        normalize(s.category).includes(q)
      );
    });
  }, [streams, query, activeCategory]);

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Browse</h1>
          <p className={styles.subtitle}>Discover live channels right now</p>
        </div>

        <input
          className={styles.search}
          placeholder="Search streams..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      <div className={styles.pills}>
        {CATEGORY_PILLS.map((p) => (
          <button
            key={p.name}
            className={`${styles.pill} ${activeCategory === p.name ? styles.pillActive : ''}`}
            onClick={() => setActiveCategory(p.name)}
          >
            <span className={styles.pillIcon}>{p.icon}</span>
            <span>{p.name}</span>
          </button>
        ))}
      </div>

      {loading ? (
        <div className={styles.grid}>
          {[...Array(6)].map((_, i) => (
            <div key={i} className={styles.skeletonCard} />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className={styles.emptyState}>
          <span className={styles.emptyIcon}>◎</span>
          <p className={styles.emptyTitle}>No streams match your filters</p>
          <p className={styles.emptySub}>Try clearing search or switching categories.</p>
          <button className={styles.clearBtn} onClick={() => { setQuery(''); setActiveCategory('All'); }}>
            Clear filters
          </button>
          <button className={styles.goLiveBtn} onClick={() => navigate('/go-live')}>
            Be the first — Go Live
          </button>
        </div>
      ) : (
        <div className={styles.grid}>
          {filtered.map((stream) => (
            <button
              key={stream.id}
              className={styles.streamCard}
              onClick={() => navigate(`/channel/${stream.username}`)}
            >
              <div className={styles.thumbnail}>
                <div className={styles.thumbnailOverlay}>
                  <span className={styles.playIcon}>▶</span>
                </div>
                <div className={styles.badges}>
                  <span className={styles.livePill}>LIVE</span>
                  <span className={styles.viewersPill}>◎ {stream.viewer_count.toLocaleString()}</span>
                </div>
              </div>

              <div className={styles.cardInfo}>
                <div className={styles.avatar}>{stream.username[0].toUpperCase()}</div>
                <div className={styles.text}>
                  <div className={styles.row}>
                    <span className={styles.streamUsername}>@{stream.username}</span>
                    <span className={styles.category}>{stream.category}</span>
                  </div>
                  <p className={styles.streamTitle}>{stream.title}</p>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

