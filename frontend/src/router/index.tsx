import { lazy, Suspense } from 'react';
import { createBrowserRouter } from 'react-router-dom';
import Login from '../pages/Login';
import Register from '../pages/Register';
import ProtectedRoute from '../components/common/ProtectedRoute';

const Dashboard = lazy(() => import('../pages/Dashboard'));

const LoadingFallback = () => (
  <div style={{
    display:        'flex',
    alignItems:     'center',
    justifyContent: 'center',
    height:         '100vh',
    background:     '#0a0a0a',
    color:          '#6b7280',
    fontFamily:     'IBM Plex Mono, monospace',
    fontSize:       '14px',
    letterSpacing:  '0.1em',
  }}>
    Loading...
  </div>
);

export const router = createBrowserRouter([
  { path: '/login',    element: <Login /> },
  { path: '/register', element: <Register /> },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: '/dashboard',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <Dashboard />
          </Suspense>
        ),
      },
    ],
  },
  { path: '/', element: <Login /> },
]);