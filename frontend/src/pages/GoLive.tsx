import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useStream } from '../hooks/useStream';
import { useAuth } from '../hooks/useAuth';
import styles from './GoLive.module.css';

const schema = z.object({
  title:    z.string().min(3, 'Title must be at least 3 characters').max(120),
  category: z.string().min(1, 'Pick a category'),
});
type FormValues = z.infer<typeof schema>;

const CATEGORIES = ['Programming','Gaming','Music','Art','DevOps','Design','Science','Sports','General'];

const OBS_STEPS = [
  { n: 1, text: 'Open OBS → Settings → Stream' },
  { n: 2, text: 'Set Service to "Custom..."' },
  { n: 3, text: 'Paste the Server URL below into "Server"' },
  { n: 4, text: 'Paste your Stream Key into "Stream Key"' },
  { n: 5, text: 'Click Apply → OK → Start Streaming' },
];

export default function GoLive() {
  const { user } = useAuth();
  const { streamKey, isLoading, error, fetchStreamKey, rotateKey, updateSettings } = useStream();
  const [keyVisible, setKeyVisible] = useState(false);
  const [copied, setCopied]         = useState<string | null>(null);
  const [saved, setSaved]           = useState(false);
  const [rotating, setRotating]     = useState(false);

  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { title: '', category: 'General' },
  });

  useEffect(() => { fetchStreamKey(); }, [fetchStreamKey]);

  const copy = async (text: string, label: string) => {
    await navigator.clipboard.writeText(text);
    setCopied(label);
    setTimeout(() => setCopied(null), 2000);
  };

  const handleRotate = async () => {
    if (!confirm('This will disconnect any active OBS session. Continue?')) return;
    setRotating(true);
    await rotateKey();
    setRotating(false);
    setKeyVisible(false);
  };

  const onSubmit = async (data: FormValues) => {
    try {
      await updateSettings(data);
      setSaved(true);
      setTimeout(() => setSaved(false), 2500);
    } catch {}
  };

  return (
    <div className={styles.page}>
      <div className={styles.pageHeader}>
        <div>
          <h1 className={styles.pageTitle}>
            <span className={styles.pageTitleDot} />
            Go Live
          </h1>
          <p className={styles.pageSubtitle}>Set up your stream and connect OBS</p>
        </div>
        <a href={`/stream/${user?.username}`} className={styles.previewLink} target="_blank">
          Preview channel →
        </a>
      </div>

      <div className={styles.grid}>
        {/* ── Left ── */}
        <div className={styles.leftCol}>
          <div className={styles.card}>
            <div className={styles.cardHeader}>
              <span className={styles.cardIcon}>◈</span>
              <h2 className={styles.cardTitle}>Stream settings</h2>
            </div>
            <form onSubmit={handleSubmit(onSubmit)} className={styles.form}>
              <div className={styles.field}>
                <label className={styles.label}>Stream title</label>
                <input
                  type="text"
                  placeholder={`${user?.username}'s stream`}
                  className={`${styles.input} ${errors.title ? styles.inputError : ''}`}
                  {...register('title')}
                />
                {errors.title && <span className={styles.fieldError}>{errors.title.message}</span>}
              </div>
              <div className={styles.field}>
                <label className={styles.label}>Category</label>
                <select className={styles.select} {...register('category')}>
                  {CATEGORIES.map(c => <option key={c} value={c}>{c}</option>)}
                </select>
              </div>
              <button type="submit" className={styles.saveBtn} disabled={isLoading}>
                {saved ? '✓ Saved' : 'Save settings'}
              </button>
            </form>
          </div>

          <div className={styles.card}>
            <div className={styles.cardHeader}>
              <span className={styles.cardIcon}>⊙</span>
              <h2 className={styles.cardTitle}>OBS Setup</h2>
            </div>
            <ol className={styles.obsList}>
              {OBS_STEPS.map(step => (
                <li key={step.n} className={styles.obsStep}>
                  <span className={styles.obsNum}>{step.n}</span>
                  <span className={styles.obsText}>{step.text}</span>
                </li>
              ))}
            </ol>
          </div>
        </div>

        {/* ── Right ── */}
        <div className={styles.rightCol}>
          <div className={styles.card}>
            <div className={styles.cardHeader}>
              <span className={styles.cardIcon}>⬡</span>
              <h2 className={styles.cardTitle}>Stream key</h2>
              <span className={styles.warnBadge}>Keep secret</span>
            </div>

            {error && <div className={styles.errorBanner}>{error}</div>}

            {isLoading && !streamKey ? (
              <div className={styles.loadingRow}><span className={styles.spinner} /> Loading...</div>
            ) : streamKey && (
              <div className={styles.keySection}>
                <div className={styles.keyField}>
                  <div className={styles.keyFieldHeader}>
                    <span className={styles.keyFieldLabel}>Server URL</span>
                    <button className={styles.copyBtn} onClick={() => copy(streamKey.rtmp_url, 'server')}>
                      {copied === 'server' ? '✓ Copied' : 'Copy'}
                    </button>
                  </div>
                  <div className={styles.keyValue}><code>{streamKey.rtmp_url}</code></div>
                </div>

                <div className={styles.keyField}>
                  <div className={styles.keyFieldHeader}>
                    <span className={styles.keyFieldLabel}>Stream Key</span>
                    <div className={styles.keyActions}>
                      <button className={styles.copyBtn} onClick={() => copy(streamKey.stream_key, 'key')}>
                        {copied === 'key' ? '✓ Copied' : 'Copy'}
                      </button>
                      <button className={styles.toggleBtn} onClick={() => setKeyVisible(v => !v)}>
                        {keyVisible ? 'Hide' : 'Show'}
                      </button>
                    </div>
                  </div>
                  <div className={styles.keyValue}>
                    <code className={styles.keyCode}>
                      {keyVisible ? streamKey.stream_key : '•'.repeat(Math.min(streamKey.stream_key.length, 40))}
                    </code>
                  </div>
                </div>

                <div className={styles.keyField}>
                  <div className={styles.keyFieldHeader}>
                    <span className={styles.keyFieldLabel}>Full OBS URL</span>
                    <button className={styles.copyBtn} onClick={() => copy(streamKey.obs_url, 'obs')}>
                      {copied === 'obs' ? '✓ Copied' : 'Copy'}
                    </button>
                  </div>
                  <div className={`${styles.keyValue} ${styles.keyValueMuted}`}>
                    <code>{keyVisible ? streamKey.obs_url : streamKey.rtmp_url + '/••••••••'}</code>
                  </div>
                </div>

                <button className={styles.rotateBtn} onClick={handleRotate} disabled={rotating}>
                  {rotating ? <><span className={styles.spinner} /> Rotating…</> : '↺ Generate new stream key'}
                </button>
                <p className={styles.rotateWarning}>Rotating your key will immediately disconnect any active stream.</p>
              </div>
            )}
          </div>

          <div className={styles.card}>
            <div className={styles.cardHeader}>
              <span className={styles.cardIcon}>◉</span>
              <h2 className={styles.cardTitle}>Stream status</h2>
            </div>
            <div className={styles.statusRow}>
              <div className={styles.statusOffline}>
                <span className={styles.statusDot} />
                <span>Offline</span>
              </div>
              <p className={styles.statusHint}>Status updates automatically when OBS connects.</p>
            </div>
            <a href={`/stream/${user?.username}`} className={styles.watchLink}>Open stream page →</a>
          </div>
        </div>
      </div>
    </div>
  );
}