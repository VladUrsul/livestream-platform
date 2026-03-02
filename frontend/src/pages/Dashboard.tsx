import { useAuth } from '../hooks/useAuth';
import styles from './Dashboard.module.css';

// Mock data — will be replaced with real API calls once stream-service is built
const featuredStreams = [
  { id: 1, username: 'techwave',    title: 'Building a Rust compiler from scratch', viewers: '2.4k', category: 'Programming', duration: '3h 12m' },
  { id: 2, username: 'pixelcraft',  title: 'Pixel art — fantasy landscape',          viewers: '891',  category: 'Art',         duration: '45m' },
  { id: 3, username: 'synthwave99', title: 'Lo-fi beats — chill Sunday session',     viewers: '5.1k', category: 'Music',       duration: '1h 20m' },
  { id: 4, username: 'cloudnative', title: 'Kubernetes deep dive — Day 3',           viewers: '1.2k', category: 'DevOps',      duration: '2h 05m' },
  { id: 5, username: 'gamedevjoe',  title: 'Making a 2D platformer in Godot',        viewers: '3.3k', category: 'GameDev',     duration: '58m' },
  { id: 6, username: 'designlabs',  title: 'UI/UX critique + redesign session',      viewers: '720',  category: 'Design',      duration: '1h 40m' },
];

const categories = [
  { name: 'Programming', icon: '⟨⟩', count: 142 },
  { name: 'Gaming',      icon: '◈',  count: 389 },
  { name: 'Music',       icon: '♩',  count: 97  },
  { name: 'Art',         icon: '◎',  count: 63  },
  { name: 'DevOps',      icon: '⊙',  count: 28  },
  { name: 'Design',      icon: '◇',  count: 44  },
];

const categoryColors: Record<string, string> = {
  Programming: '#3b82f6',
  Art:         '#a855f7',
  Music:       '#ec4899',
  DevOps:      '#22c55e',
  GameDev:     '#f97316',
  Design:      '#e8ff47',
};

export default function Dashboard() {
  const { user } = useAuth();

  const hour = new Date().getHours();
  const greeting = hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening';

  return (
    <div className={styles.page}>

      {/* Hero greeting */}
      <div className={styles.hero}>
        <div className={styles.heroText}>
          <p className={styles.heroGreeting}>{greeting}</p>
          <h1 className={styles.heroTitle}>
            Welcome back,{' '}
            <span className={styles.heroUsername}>@{user?.username}</span>
          </h1>
          <p className={styles.heroSub}>
            <span className={styles.liveCount}>◉ 847 streams</span> live right now
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

      {/* Categories */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Browse by category</h2>
          <a href="/browse" className={styles.sectionLink}>See all →</a>
        </div>
        <div className={styles.categoryGrid}>
          {categories.map((cat) => (
            <a key={cat.name} href={`/browse/${cat.name.toLowerCase()}`} className={styles.categoryCard}>
              <span className={styles.categoryIcon}>{cat.icon}</span>
              <span className={styles.categoryName}>{cat.name}</span>
              <span className={styles.categoryCount}>{cat.count} live</span>
            </a>
          ))}
        </div>
      </section>

      {/* Live streams */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>
            Live now
            <span className={styles.liveBadge}>
              <span className={styles.liveDot} />
              LIVE
            </span>
          </h2>
          <a href="/live" className={styles.sectionLink}>See all →</a>
        </div>

        <div className={styles.streamGrid}>
          {featuredStreams.map((stream) => (
            <a key={stream.id} href={`/stream/${stream.username}`} className={styles.streamCard}>
              {/* Thumbnail placeholder */}
              <div className={styles.streamThumbnail}>
                <div
                  className={styles.thumbnailBg}
                  style={{ background: `linear-gradient(135deg, ${categoryColors[stream.category] ?? '#374151'}22, #111)` }}
                />
                <div className={styles.thumbnailOverlay}>
                  <span className={styles.thumbnailIcon}>▶</span>
                </div>
                <div className={styles.streamBadges}>
                  <span className={styles.livePill}>
                    <span className={styles.livePillDot} />
                    LIVE
                  </span>
                  <span className={styles.viewersPill}>◎ {stream.viewers}</span>
                </div>
                <span className={styles.durationPill}>{stream.duration}</span>
              </div>

              {/* Stream info */}
              <div className={styles.streamCardInfo}>
                <div className={styles.streamCardAvatar}>
                  {stream.username[0].toUpperCase()}
                </div>
                <div className={styles.streamCardText}>
                  <p className={styles.streamCardTitle}>{stream.title}</p>
                  <p className={styles.streamCardUsername}>@{stream.username}</p>
                  <span
                    className={styles.streamCardCategory}
                    style={{ borderColor: `${categoryColors[stream.category] ?? '#374151'}44`, color: categoryColors[stream.category] ?? '#6b7280' }}
                  >
                    {stream.category}
                  </span>
                </div>
              </div>
            </a>
          ))}
        </div>
      </section>

    </div>
  );
}