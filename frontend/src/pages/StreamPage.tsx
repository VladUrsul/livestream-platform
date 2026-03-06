import { useEffect, useRef, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import Hls from 'hls.js';
import { streamService } from '../services/streamService';
import { userService } from '../services/userService';
import { useAuth } from '../hooks/useAuth';
import { type StreamInfo } from '../types/stream.types';
import ChatPanel from '../components/chat/ChatPanel';
import styles from './StreamPage.module.css';

const POLL_INTERVAL = 5000;

export default function StreamPage() {
  const { username } = useParams<{ username: string }>();
  const navigate     = useNavigate();
  const { user: me } = useAuth();

  const videoRef = useRef<HTMLVideoElement>(null);
  const hlsRef   = useRef<Hls | null>(null);

  const [stream,     setStream]     = useState<StreamInfo | null>(null);
  const [loading,    setLoading]    = useState(true);
  const [error,      setError]      = useState<string | null>(null);
  const [viewers,    setViewers]    = useState(0);
  const [muted,      setMuted]      = useState(true);
  const [volume,     setVolume]     = useState(1);
  const [fullscreen, setFullscreen] = useState(false);
  const [following,  setFollowing]  = useState(false);
  const [followLoading, setFollowLoading] = useState(false);

  const isOwner = me?.username === username;

  // ── HLS ───────────────────────────────────────────────────────────
  const destroyHls = useCallback(() => {
    if (hlsRef.current) { hlsRef.current.destroy(); hlsRef.current = null; }
    if (videoRef.current) { videoRef.current.removeAttribute('src'); videoRef.current.load(); }
  }, []);

  const startHls = useCallback((hlsUrl: string) => {
    const attemptStart = (attemptsLeft: number) => {
      const video = videoRef.current;
      if (!video) {
        if (attemptsLeft > 0) { setTimeout(() => attemptStart(attemptsLeft - 1), 100); return; }
        return;
      }
      destroyHls();
      if (Hls.isSupported()) {
        const hls = new Hls({
          lowLatencyMode: true,
          backBufferLength: 30,
          maxBufferLength: 10,
          manifestLoadingMaxRetry: 6,
          manifestLoadingRetryDelay: 1000,
          levelLoadingMaxRetry: 6,
          fragLoadingMaxRetry: 6,
        });
        hls.on(Hls.Events.MANIFEST_PARSED, () => {
          video.muted = true;
          video.play().catch(() => {});
        });
        hls.on(Hls.Events.ERROR, (_, data) => {
          if (data.fatal) {
            if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
              setTimeout(() => {
                if (hlsRef.current) { hls.loadSource(hlsUrl); hls.startLoad(); }
              }, 2000);
            } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
              hls.recoverMediaError();
            } else {
              setError('Playback error');
              destroyHls();
            }
          } else if (data.details === Hls.ErrorDetails.BUFFER_APPEND_ERROR) {
            hls.recoverMediaError();
          }
        });
        hls.loadSource(hlsUrl);
        hls.attachMedia(video);
        hlsRef.current = hls;
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = hlsUrl;
        video.play().catch(() => {});
      }
    };
    attemptStart(10);
  }, [destroyHls]);

  // ── Poll stream + init follow state ───────────────────────────────
  useEffect(() => {
    if (!username) return;
    let cancelled  = false;
    let lastHlsUrl: string | null = null;

    const init = async () => {
      try {
        const [info, isFollowingData] = await Promise.all([
          streamService.getStreamInfo(username),
          me && me.username !== username
            ? userService.isFollowing(username)
            : Promise.resolve(false),
        ]);
        if (cancelled) return;
        setStream(info);
        setViewers(info.viewer_count);
        setFollowing(isFollowingData);
        setLoading(false);
        if (info.status === 'live' && info.hls_url) {
          lastHlsUrl = info.hls_url;
          startHls(info.hls_url);
        }
      } catch (err: any) {
        if (!cancelled) { setError(err.response?.data?.error || 'Stream not found'); setLoading(false); }
      }
    };

    init();
    streamService.joinStream(username).catch(() => {});

    const interval = setInterval(async () => {
      try {
        const info = await streamService.getStreamInfo(username);
        if (cancelled) return;
        setStream(info);
        setViewers(info.viewer_count);
        if (info.status === 'live' && info.hls_url && info.hls_url !== lastHlsUrl) {
          lastHlsUrl = info.hls_url;
          startHls(info.hls_url);
        } else if (info.status !== 'live' && lastHlsUrl) {
          lastHlsUrl = null;
          destroyHls();
        }
      } catch {}
    }, POLL_INTERVAL);

    return () => {
      cancelled = true;
      clearInterval(interval);
      streamService.leaveStream(username).catch(() => {});
      destroyHls();
    };
  }, [username, me, startHls, destroyHls]);

  // ── Follow ────────────────────────────────────────────────────────
  const toggleFollow = async () => {
    if (!me || isOwner || !username) return;
    setFollowLoading(true);
    try {
      if (following) {
        await userService.unfollow(username);
        setFollowing(false);
      } else {
        await userService.follow(username);
        setFollowing(true);
      }
    } catch {}
    setFollowLoading(false);
  };

  // ── Controls ──────────────────────────────────────────────────────
  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = parseFloat(e.target.value);
    setVolume(v);
    if (videoRef.current) videoRef.current.volume = v;
    setMuted(v === 0);
  };

  const toggleMute = () => {
    if (!videoRef.current) return;
    const next = !muted;
    setMuted(next);
    videoRef.current.muted = next;
  };

  const toggleFullscreen = () => {
    const el = document.getElementById('stream-player-wrap');
    if (!document.fullscreenElement) {
      el?.requestFullscreen();
      setFullscreen(true);
    } else {
      document.exitFullscreen();
      setFullscreen(false);
    }
  };

  // ── Render ────────────────────────────────────────────────────────
  if (!username) return (
    <div className={styles.centerState}>
      <span className={styles.errorIcon}>◎</span>
      <p>Channel not found</p>
    </div>
  );

  if (loading) return (
    <div className={styles.centerState}>
      <span className={styles.spinner} />
      <span>Loading stream...</span>
    </div>
  );

  if (error && !stream) return (
    <div className={styles.centerState}>
      <span className={styles.errorIcon}>◎</span>
      <p>{error}</p>
    </div>
  );

  const isLive = stream?.status === 'live';

  return (
    <div className={styles.page}>
      <div className={styles.layout}>

        {/* ── Player column ── */}
        <div className={styles.playerCol}>
          <div className={styles.playerWrap} id="stream-player-wrap">

            <video
              ref={videoRef}
              className={styles.video}
              style={{ opacity: isLive ? 1 : 0, position: isLive ? 'relative' : 'absolute' }}
              playsInline
              autoPlay
            />

            {!isLive && (
              <div className={styles.offlineScreen}>
                <div className={styles.offlineIcon}>◈</div>
                <h3 className={styles.offlineTitle}>@{username} is offline</h3>
                <p className={styles.offlineText}>
                  {stream?.status === 'ended' ? 'This stream has ended.' : 'Waiting for stream to start...'}
                </p>
                <p className={styles.offlinePoll}>Checking every 5 seconds</p>
              </div>
            )}

            {isLive && muted && (
              <button
                className={styles.unmutePrompt}
                onClick={() => { if (videoRef.current) videoRef.current.muted = false; setMuted(false); }}
              >
                <span className={styles.unmuteIcon}>🔇</span>
                <span>Click to unmute</span>
              </button>
            )}

            {isLive && (
              <>
                <div className={styles.controls}>
                  <div className={styles.controlsLeft}>
                    <button className={styles.controlBtn} onClick={toggleMute}>
                      {muted || volume === 0 ? '🔇' : volume < 0.5 ? '🔉' : '🔊'}
                    </button>
                    <input
                      type="range" min="0" max="1" step="0.05"
                      value={muted ? 0 : volume}
                      onChange={handleVolumeChange}
                      className={styles.volumeSlider}
                    />
                  </div>
                  <div className={styles.controlsRight}>
                    <button className={styles.controlBtn} onClick={toggleFullscreen}>
                      {fullscreen ? '⊡' : '⊞'}
                    </button>
                  </div>
                </div>
                <div className={styles.liveBadge}><span className={styles.liveDot} />LIVE</div>
                <div className={styles.viewerBadge}>◎ {viewers.toLocaleString()}</div>
              </>
            )}
          </div>

          {/* Info bar */}
          {stream && (
            <div className={styles.infoBar}>
              <div className={styles.infoBarLeft}>
                <button
                  className={styles.streamerAvatar}
                  onClick={() => navigate(`/channel/${username}`)}
                >
                  {username?.[0]?.toUpperCase()}
                </button>
                <div className={styles.infoBarText}>
                  <h1 className={styles.streamTitle}>{stream.title}</h1>
                  <div className={styles.streamMeta}>
                    <span className={styles.streamUsername}>@{stream.username}</span>
                    <span className={styles.metaDot}>·</span>
                    <span className={styles.streamCategory}>{stream.category}</span>
                    {isLive && stream.started_at && (
                      <>
                        <span className={styles.metaDot}>·</span>
                        <span className={styles.streamDuration}>
                          Started {new Date(stream.started_at).toLocaleTimeString([], {
                            hour: '2-digit', minute: '2-digit',
                          })}
                        </span>
                      </>
                    )}
                  </div>
                </div>
              </div>
              <div className={styles.infoBarRight}>
                {isLive && (
                  <div className={styles.viewerCount}>
                    <span className={styles.viewerDot} />
                    <span>{viewers.toLocaleString()} watching</span>
                  </div>
                )}
                {!isOwner && me && (
                  <button
                    className={`${styles.followBtn} ${following ? styles.followBtnActive : ''}`}
                    onClick={toggleFollow}
                    disabled={followLoading}
                  >
                    {followLoading ? '...' : following ? 'Unfollow' : 'Follow'}
                  </button>
                )}
              </div>
            </div>
          )}
        </div>

        {/* ── Chat column ── */}
        <div className={styles.chatCol}>
          {username && (
            <ChatPanel roomID={username} isOwner={isOwner} />
          )}
        </div>

      </div>
    </div>
  );
}