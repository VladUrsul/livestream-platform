import { lazy, Suspense } from 'react';
import { createBrowserRouter, Navigate } from 'react-router-dom';
import Login from '../pages/Login';
import Register from '../pages/Register';
import ProtectedRoute from '../components/common/ProtectedRoute';
import Layout from '../components/common/Layout';

const Dashboard = lazy(() => import('../pages/Dashboard'));
const GoLive      = lazy(() => import('../pages/GoLive'));
const StreamPage  = lazy(() => import('../pages/StreamPage'));

const Loader = () => (
  <div style={{
    display: 'flex', alignItems: 'center', justifyContent: 'center',
    height: '100vh', background: '#0a0a0a', color: '#6b7280',
    fontFamily: 'IBM Plex Mono, monospace', fontSize: '14px',
  }}>
    ◈ &nbsp; loading
  </div>
);

const S = (C: React.ComponentType) => (
  <Suspense fallback={<Loader />}><C /></Suspense>
);

export const router = createBrowserRouter([
  // Public routes
  { path: '/login',    element: <Login /> },
  { path: '/register', element: <Register /> },

  // Protected routes — all share the Layout (navbar + sidebar)
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <Layout />,
        children: [
          {
            path: '/dashboard',
            element: <Suspense fallback={<Loader />}><Dashboard /></Suspense>,
          },
          { path: '/go-live',          
            element: S(GoLive) 
          },
          { path: '/stream/:username', 
            element: S(StreamPage) 
          },
        ],
      },
    ],
  },

  { path: '/', element: <Navigate to="/login" replace /> },
]);