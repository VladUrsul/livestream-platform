import { useAuth } from '../hooks/useAuth';

export default function Dashboard() {
  const { user, logout } = useAuth();

  return (
    <div style={{
      minHeight: '100vh', background: '#0a0a0a', color: '#f9fafb',
      fontFamily: 'Syne, sans-serif', display: 'flex',
      flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 24,
    }}>
      <h1 style={{ fontSize: 32, fontWeight: 800, margin: 0 }}>
        Welcome, @{user?.username}
      </h1>
      <p style={{ color: '#6b7280', margin: 0 }}>
        Dashboard coming soon — auth is working!
      </p>
      <button
        onClick={logout}
        style={{
          padding: '10px 24px', background: 'transparent',
          border: '1px solid #374151', borderRadius: 8,
          color: '#9ca3af', cursor: 'pointer',
          fontFamily: 'Syne, sans-serif', fontSize: 14,
        }}
      >
        Sign out
      </button>
    </div>
  );
}