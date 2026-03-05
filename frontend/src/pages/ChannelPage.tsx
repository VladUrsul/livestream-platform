import { useEffect, useRef, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Hls from 'hls.js';
import { userService } from '../services/userService';
import { streamService } from '../services/streamService';
import { useAuth } from '../hooks/useAuth';
import { type Profile } from '../types/user.types';
import { type StreamInfo } from '../types/stream.types';
import styles from './ChannelPage.module.css';

export default function ChannelPage() {
  const { username } = useParams<{ username: string }>();
  const { user: me } = useAuth();
  const navigate = useNavigate();

  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef   = useRef<Hls | null>(null);

  const [profile,     setProfile]     = useState<Profile | null>(null);
  const [stream,      setStream]      = useState<StreamInfo | null>(null);
  const [loading,     setLoading]     = useState(true);
  const [error,       setError]       = useState<string | null>(null);
  const [following,   setFollowing]   = useState(false);
  const [followLoading, setFollowLoading] = useState(false);
  const [muted,       setMuted]       = useState(true);
  const [activeTab,   setActiveTab]   = useState<'about' | 'schedule'>('about');

  const isOwner = me?.username === username;

  // ── HLS ──────────────────────────────────────────────────────────
  const destroyHls = useCallback(() => {
    if (hlsRef.current) { hlsRef.current.destroy(); hlsRef.current = null; }
    if (videoRef.current) { videoRef.current.removeAttribute('src'); videoRef.current.load(); }
  }, []);

  const startHls = useCallback((hlsUrl: string) => {
    const attempt = (tries: number) => {
      const video = videoRef.current;
      if (!video) {
        if (tries > 0) { setTimeout(() => attempt(tries - 1), 100); return; }
        return;
      }
      destroyHls();
      if (Hls.isSupported()) {
        const hls = new Hls({
          lowLatencyMode: true,
          manifestLoadingMaxRetry: 8,
          manifestLoadingRetryDelay: 1500,
        });
        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          video.muted = true;
          video.play().catch(() => {});
        });
        hls.on(Hls.Events.ERROR, (_, data) => {
          if (data.fatal) {
            if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
              setTimeout(() => { if (hlsRef.current) { hls.loadSource(hlsUrl); hls.startLoad(); } }, 2000);
            } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
              hls.recoverMediaError();
            } else {
              destroyHls();
            }
          }
        });
        hls.loadSource(hlsUrl);
        hls.attachMedia(video);
        hlsRef.current = hls;
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = hlsUrl;
        video.muted = true;
        video.play().catch(() => {});
      }
    };
    attempt(10);
  }, [destroyHls]);

  // ── Load profile + stream ─────────────────────────────────────────
  useEffect(() => {
    if (!username) return;
    let cancelled = false;
    let lastHlsUrl: string | null = null;

    const load = async () => {
      try {
        const [profileData, streamData, isFollowingData] = await Promise.all([
          userService.getProfile(username),
          streamService.getStreamInfo(username).catch(() => null),
          me && me.username !== username
          ? userService.isFollowing(username)
          : Promise.resolve(false),
        ]);
        if (cancelled) return;
        setProfile(profileData);
        setStream(streamData);
        setFollowing(isFollowingData);
        setLoading(false);

        if (streamData?.status === 'live' && streamData.hls_url && streamData.hls_url !== lastHlsUrl) {
          lastHlsUrl = streamData.hls_url;
          startHls(streamData.hls_url);
        }
      } catch {
        if (!cancelled) { setError('Channel not found'); setLoading(false); }
      }
    };

    load();

    // Poll stream status every 6s
    const poll = setInterval(async () => {
      try {
        const streamData = await streamService.getStreamInfo(username);
        if (cancelled) return;
        setStream(streamData);
        if (streamData.status === 'live' && streamData.hls_url && streamData.hls_url !== lastHlsUrl) {
          lastHlsUrl = streamData.hls_url;
          startHls(streamData.hls_url);
        } else if (streamData.status !== 'live' && lastHlsUrl) {
          lastHlsUrl = null;
          destroyHls();
        }
      } catch {}
    }, 6000);

    return () => {
      cancelled = true;
      clearInterval(poll);
      destroyHls();
    };
  }, [username, startHls, destroyHls]);

  // ── Follow ────────────────────────────────────────────────────────
  const toggleFollow = async () => {
    if (!me || isOwner) return;
    setFollowLoading(true);
    try {
      if (following) {
        await userService.unfollow(username!);
        setFollowing(false);
        setProfile(p => p ? { ...p, followers: Math.max(0, p.followers - 1) } : p);
      } else {
        await userService.follow(username!);
        setFollowing(true);
        setProfile(p => p ? { ...p, followers: p.followers + 1 } : p);
      }
    } catch {}
    setFollowLoading(false);
  };

  if (loading) return (
    <div className={styles.centerState}>
      <span className={styles.spinner} />
    </div>
  );

  if (error || !profile) return (
    <div className={styles.centerState}>
      <span className={styles.errorIcon}>◎</span>
      <p>{error || 'Channel not found'}</p>
    </div>
  );

  const isLive = stream?.status === 'live';
  const avatarLetter = (profile.display_name || profile.username)[0].toUpperCase();

  return (
    <div className={styles.page}>

      {/* ── Banner + player area ── */}
      <div className={styles.hero}>
        <div className={styles.playerWrap} id="channel-player">

          {/* Always-mounted video */}
          <video
            ref={videoRef}
            className={styles.video}
            style={{ opacity: isLive ? 1 : 0 }}
            playsInline
            autoPlay
          />

          {/* Offline overlay */}
          {!isLive && (
            <div className={styles.offlineOverlay}>
              <div className={styles.offlineAvatar}>
                {profile.avatar_url
                  ? <img src={profile.avatar_url} alt={profile.username} />
                  : <span>{avatarLetter}</span>
                }
              </div>
              <p className={styles.offlineText}>
                {stream?.status === 'ended' ? 'Stream ended' : `@${username} is offline`}
              </p>
            </div>
          )}

          {/* Live badges */}
          {isLive && (
            <>
              <div className={styles.liveBadge}><span className={styles.liveDot} />LIVE</div>
              <div className={styles.viewerBadge}>◎ {stream?.viewer_count?.toLocaleString()}</div>
              <button
                className={styles.unmuteBtn}
                style={{ display: muted ? 'flex' : 'none' }}
                onClick={() => { if (videoRef.current) videoRef.current.muted = false; setMuted(false); }}
              >
                🔇 Click to unmute
              </button>
            </>
          )}
        </div>
      </div>

      {/* ── Profile header ── */}
      <div className={styles.profileHeader}>
        <div className={styles.profileLeft}>
          <div className={styles.avatar}>
            {profile.avatar_url
              ? <img src={profile.avatar_url} alt={profile.username} />
              : <span>{avatarLetter}</span>
            }
            {isLive && <span className={styles.avatarLiveDot} />}
          </div>
          <div className={styles.profileInfo}>
            <div className={styles.profileNameRow}>
              <h1 className={styles.displayName}>
                {profile.display_name || profile.username}
              </h1>
              {isLive && <span className={styles.liveTag}>LIVE</span>}
            </div>
            <p className={styles.usernameTag}>@{profile.username}</p>
            <div className={styles.stats}>
              <span className={styles.stat}>
                <strong>{profile.followers.toLocaleString()}</strong> followers
              </span>
              <span className={styles.statDot}>·</span>
              <span className={styles.stat}>
                <strong>{profile.following.toLocaleString()}</strong> following
              </span>
            </div>
          </div>
        </div>

        <div className={styles.profileActions}>
          {isOwner ? (
            <button className={styles.editBtn} onClick={() => navigate('/go-live')}>
              ◈ Go Live
            </button>
          ) : me ? (
            <button
              className={`${styles.followBtn} ${following ? styles.followBtnActive : ''}`}
              onClick={toggleFollow}
              disabled={followLoading}
            >
              {followLoading ? '...' : following ? 'Following' : 'Follow'}
            </button>
          ) : null}
        </div>
      </div>

      {/* ── Stream title (when live) ── */}
      {isLive && stream && (
        <div className={styles.streamTitle}>
          <span className={styles.streamTitleDot} />
          <span>{stream.title}</span>
          <span className={styles.streamCategory}>{stream.category}</span>
        </div>
      )}

      {/* ── Tabs ── */}
      <div className={styles.tabs}>
        <button
          className={`${styles.tab} ${activeTab === 'about' ? styles.tabActive : ''}`}
          onClick={() => setActiveTab('about')}
        >
          About
        </button>
        <button
          className={`${styles.tab} ${activeTab === 'schedule' ? styles.tabActive : ''}`}
          onClick={() => setActiveTab('schedule')}
        >
          Schedule
        </button>
      </div>

      {/* ── Tab content ── */}
      <div className={styles.tabContent}>
        {activeTab === 'about' && (
          <div className={styles.aboutSection}>
            {profile.bio ? (
              <p className={styles.bio}>{profile.bio}</p>
            ) : (
              <p className={styles.bioEmpty}>
                {isOwner ? 'Add a bio on your profile settings.' : 'No bio yet.'}
              </p>
            )}
            <div className={styles.aboutMeta}>
              <span className={styles.aboutMetaItem}>
                ◈ Joined {new Date(profile.created_at).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
              </span>
            </div>
          </div>
        )}
        {activeTab === 'schedule' && (
          <div className={styles.aboutSection}>
            <p className={styles.bioEmpty}>No schedule set.</p>
          </div>
        )}
      </div>

    </div>
  );
}