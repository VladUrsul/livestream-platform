import { useEffect, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import Hls from 'hls.js';
import { streamService } from '../services/streamService';
import { type StreamInfo } from '../types/stream.types';
import styles from './StreamPage.module.css';

export default function StreamPage() {
  const { username } = useParams<{ username: string }>();
  const videoRef    = useRef<HTMLVideoElement>(null);
  const hlsRef      = useRef<Hls | null>(null);

  const [stream,   setStream]   = useState<StreamInfo | null>(null);
  const [loading,  setLoading]  = useState(true);
  const [error,    setError]    = useState<string | null>(null);
  const [viewers,  setViewers]  = useState(0);
  const [muted,    setMuted]    = useState(false);
  const [volume,   setVolume]   = useState(1);
  const [fullscreen, setFullscreen] = useState(false);

  // Fetch stream info and join
  useEffect(() => {
    if (!username) return;
    let cancelled = false;

    const load = async () => {
      try {
        const info = await streamService.getStreamInfo(username);
        if (!cancelled) {
          setStream(info);
          setViewers(info.viewer_count);
        }
        const join = await streamService.joinStream(username);
      } catch (err: any) {
        if (!cancelled) setError(err.response?.data?.error || 'Stream not found');
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    load();

    // Poll stream status every 10 seconds
    const poll = setInterval(async () => {
      try {
        const info = await streamService.getStreamInfo(username);
        if (!cancelled) {
          setStream(info);
          setViewers(info.viewer_count);
        }
      } catch {}
    }, 10_000);

    return () => {
      cancelled = true;
      clearInterval(poll);
      streamService.leaveStream(username).catch(() => {});
    };
  }, [username]);

  // Set up HLS player when stream goes live
  useEffect(() => {
    if (!stream?.hls_url || !videoRef.current) return;

    const video = videoRef.current;

    if (Hls.isSupported()) {
      const hls = new Hls({
        lowLatencyMode:      true,
        backBufferLength:    30,
        maxBufferLength:     10,
        maxMaxBufferLength:  20,
      });
      hls.loadSource(stream.hls_url);
      hls.attachMedia(video);
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        video.play().catch(() => {});
      });
      hls.on(Hls.Events.ERROR, (_, data) => {
        if (data.fatal) {
          setError('Stream playback error — please refresh');
        }
      });
      hlsRef.current = hls;
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      // Native HLS (Safari)
      video.src = stream.hls_url;
      video.play().catch(() => {});
    }

    return () => {
      hlsRef.current?.destroy();
      hlsRef.current = null;
    };
  }, [stream?.hls_url, stream?.status]);

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

  if (loading) return (
    <div className={styles.centerState}>
      <span className={styles.spinner} />
      <span>Loading stream...</span>
    </div>
  );

  if (error || !stream) return (
    <div className={styles.centerState}>
      <span className={styles.errorIcon}>◎</span>
      <p>{error || 'Stream not found'}</p>
    </div>
  );

  const isLive = stream.status === 'live';

  return (
    <div className={styles.page}>
      <div className={styles.layout}>

        {/* ── Player column ── */}
        <div className={styles.playerCol}>

          {/* Video player */}
          <div className={styles.playerWrap} id="stream-player-wrap">
            {isLive ? (
              <>
                <video
                  ref={videoRef}
                  className={styles.video}
                  playsInline
                  autoPlay
                />
                {/* Controls overlay */}
                <div className={styles.controls}>
                  <div className={styles.controlsLeft}>
                    <button className={styles.controlBtn} onClick={toggleMute} title="Mute">
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
                    <button className={styles.controlBtn} onClick={toggleFullscreen} title="Fullscreen">
                      {fullscreen ? '⊡' : '⊞'}
                    </button>
                  </div>
                </div>
                {/* Live badge */}
                <div className={styles.liveBadge}>
                  <span className={styles.liveDot} />
                  LIVE
                </div>
                {/* Viewer count */}
                <div className={styles.viewerBadge}>◎ {viewers.toLocaleString()}</div>
              </>
            ) : (
              <div className={styles.offlineScreen}>
                <div className={styles.offlineIcon}>◈</div>
                <h3 className={styles.offlineTitle}>@{username} is offline</h3>
                <p className={styles.offlineText}>
                  {stream.status === 'ended'
                    ? 'This stream has ended.'
                    : 'The streamer has not gone live yet.'}
                </p>
              </div>
            )}
          </div>

          {/* Stream info bar */}
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
                        Started {new Date(stream.started_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
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
        </div>
      </div>
    </div>
  );
}