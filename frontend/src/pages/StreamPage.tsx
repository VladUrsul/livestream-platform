import { useEffect, useRef, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import Hls from 'hls.js';
import { streamService } from '../services/streamService';
import { type StreamInfo } from '../types/stream.types';
import styles from './StreamPage.module.css';

const POLL_INTERVAL = 5000; // poll every 5s so going-live is detected quickly

export default function StreamPage() {
  const { username } = useParams<{ username: string }>();
  const videoRef  = useRef<HTMLVideoElement>(null);
  const hlsRef    = useRef<Hls | null>(null);

  const [stream,     setStream]     = useState<StreamInfo | null>(null);
  const [loading,    setLoading]    = useState(true);
  const [error,      setError]      = useState<string | null>(null);
  const [viewers,    setViewers]    = useState(0);
  const [muted,      setMuted]      = useState(true);
  const [volume,     setVolume]     = useState(1);
  const [fullscreen, setFullscreen] = useState(false);

  // ── Destroy HLS instance ──────────────────────────────────────────
  const destroyHls = useCallback(() => {
    if (hlsRef.current) {
      hlsRef.current.destroy();
      hlsRef.current = null;
    }
    if (videoRef.current) {
      videoRef.current.removeAttribute('src');
      videoRef.current.load();
    }
  }, []);

  // ── Start HLS playback for a given URL ───────────────────────────
  const startHls = useCallback((hlsUrl: string) => {
    const attemptStart = (attemptsLeft: number) => {
      const video = videoRef.current;

      if (!video) {
        if (attemptsLeft > 0) {
          // Video element not in DOM yet — retry after next frame
          setTimeout(() => attemptStart(attemptsLeft - 1), 100);
          return;
        }
        console.warn('[HLS] video element never became available');
        return;
      }

      console.log('[HLS] startHls attached', { hlsUrl });
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
          console.log('[HLS] manifest parsed — playing');
          video.muted = true; // ensure muted for autoplay policy
          video.play().catch(e => console.warn('[HLS] play failed', e));
        });

        hls.on(Hls.Events.ERROR, (_, data) => {
          console.error('[HLS] error', data.type, data.details, data.fatal);
          if (data.fatal) {
            if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
              // Manifest may not exist yet — keep retrying every 2s
              setTimeout(() => {
                if (hlsRef.current) {
                  hls.loadSource(hlsUrl);
                  hls.startLoad();
                }
              }, 2000);
            } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
              hls.recoverMediaError();
            } else {
              setError('Playback error');
              destroyHls();
            }
          } else if (data.details === Hls.ErrorDetails.BUFFER_APPEND_ERROR) {
            // Non-fatal buffer errors — recover media
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

    attemptStart(10); // retry up to 10 times × 100ms = 1 second max wait
  }, [destroyHls]);

  // ── Poll stream info ──────────────────────────────────────────────
  useEffect(() => {
    if (!username) return;
    let cancelled = false;
    // Track what HLS URL we last loaded so we only re-init when it changes
    let lastHlsUrl: string | null = null;
    let lastStatus: string | null = null;

    const fetchAndUpdate = async () => {
      try {
        const info = await streamService.getStreamInfo(username);
        if (cancelled) return;

        setStream(info);
        setViewers(info.viewer_count);
        setError(null);

        const hlsUrl = info.hls_url ?? null;
        const status = info.status;

        if (status === 'live' && hlsUrl) {
          // Start or restart player if URL changed (new stream session)
          if (hlsUrl !== lastHlsUrl) {
            lastHlsUrl = hlsUrl;
            startHls(hlsUrl);
          }
        } else {
          // Stream went offline — destroy player
          if (lastHlsUrl !== null) {
            lastHlsUrl = null;
            destroyHls();
          }
        }

        lastStatus = status;
      } catch (err: any) {
        if (!cancelled) {
          setError(err.response?.data?.error || 'Stream not found');
          setLoading(false);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    // Initial fetch immediately
    fetchAndUpdate();

    // Join viewer count
    streamService.joinStream(username).catch(() => {});

    const interval = setInterval(fetchAndUpdate, POLL_INTERVAL);

    return () => {
      cancelled = true;
      clearInterval(interval);
      streamService.leaveStream(username).catch(() => {});
      destroyHls();
    };
  }, [username, startHls, destroyHls]);

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
        <div className={styles.playerCol}>

          {/* Video player */}
          <div className={styles.playerWrap} id="stream-player-wrap">

            {/* Video is ALWAYS in the DOM — display toggled via CSS only */}
            <video
              ref={videoRef}
              className={styles.video}
              style={{ opacity: isLive ? 1 : 0, position: isLive ? 'relative' : 'absolute' }}
              playsInline
              autoPlay
            />

            {/* Offline screen sits on top when not live */}
            {!isLive && (
              <div className={styles.offlineScreen}>
                <div className={styles.offlineIcon}>◈</div>
                <h3 className={styles.offlineTitle}>@{username} is offline</h3>
                <p className={styles.offlineText}>
                  {stream?.status === 'ended'
                    ? 'This stream has ended.'
                    : 'Waiting for stream to start...'}
                </p>
                <p className={styles.offlinePoll}>Checking every 5 seconds</p>
              </div>
            )}

            {/* Unmute prompt — shown when playing muted after autoplay */}
            {isLive && muted && (
              <button
                className={styles.unmutePrompt}
                onClick={() => {
                  if (videoRef.current) videoRef.current.muted = false;
                  setMuted(false);
                }}
              >
                <span className={styles.unmuteIcon}>🔇</span>
                <span>Click to unmute</span>
              </button>
            )}

            {/* Overlays — only when live */}
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

                <div className={styles.liveBadge}>
                  <span className={styles.liveDot} />
                  LIVE
                </div>

                <div className={styles.viewerBadge}>
                  ◎ {viewers.toLocaleString()}
                </div>
              </>
            )}
          </div>

          {/* Info bar */}
          {stream && (
            <div className={styles.infoBar}>
              <div className={styles.infoBarLeft}>
                <div className={styles.streamerAvatar}>
                  {username?.[0]?.toUpperCase()}
                </div>
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
                <button className={styles.followBtn}>Follow</button>
              </div>
            </div>
          )}

        </div>
      </div>
    </div>
  );
}