import { useEffect } from 'react';
import { RouterProvider } from 'react-router-dom';
import { Provider, useDispatch, useSelector } from 'react-redux';
import { 
  store, 
  type RootState, 
  type AppDispatch 
} from './store';
import { router } from './router';
import { getMe } from './store/authSlice';

// AppInit runs once on mount. If a token exists in localStorage it calls
// getMe to validate it and load fresh user data from the server.
function AppInit() {
  const dispatch = useDispatch<AppDispatch>();
  const { accessToken, isLoading } = useSelector((state: RootState) => state.auth);

  useEffect(() => {
    if (accessToken) {
      dispatch(getMe());
    } else {
      // No token — mark loading as done immediately
      dispatch({ type: 'auth/getMe/rejected' });
    }
  }, []);

  // Show a blank loading screen while we validate the token.
  // This prevents the login page from flashing before redirect.
  if (isLoading) {
    return (
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
        ◈ &nbsp; loading
      </div>
    );
  }

  return <RouterProvider router={router} />;
}

export default function App() {
  return (
    <Provider store={store}>
      <AppInit />
    </Provider>
  );
}