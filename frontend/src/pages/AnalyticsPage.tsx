import styles from './AnalyticsPage.module.css';

export default function AnalyticsPage() {
  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Analytics</h1>
          <p className={styles.subtitle}>Metrics will appear here after you go live.</p>
        </div>
      </div>

      <div className={styles.grid}>
        <div className={styles.card}>
          <div className={styles.cardLabel}>Total views</div>
          <div className={styles.cardValue}>—</div>
          <div className={styles.cardHint}>Coming soon</div>
        </div>

        <div className={styles.card}>
          <div className={styles.cardLabel}>Avg. watch time</div>
          <div className={styles.cardValue}>—</div>
          <div className={styles.cardHint}>Coming soon</div>
        </div>

        <div className={styles.card}>
          <div className={styles.cardLabel}>New followers</div>
          <div className={styles.cardValue}>—</div>
          <div className={styles.cardHint}>Coming soon</div>
        </div>
      </div>

      <div className={styles.notice}>
        This page is currently frontend-only. When backend analytics endpoints are ready, this view will
        switch from placeholders to real data.
      </div>
    </div>
  );
}

