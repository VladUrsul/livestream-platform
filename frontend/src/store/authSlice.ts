import { 
  createSlice, 
  createAsyncThunk, 
  type PayloadAction 
} from '@reduxjs/toolkit';
import { 
  type AuthState, 
  type LoginInput, 
  type RegisterInput 
} from '../types/auth.types';
import { authService } from '../services/authService';

const STORAGE_KEYS = {
  ACCESS_TOKEN:  'ls_access_token',
  REFRESH_TOKEN: 'ls_refresh_token',
};

const clearStorage = () => {
  localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN);
  localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
};

// ── Async Thunks ──────────────────────────────────────────────────────────────

// Called once on app load to validate the stored token and get fresh user data
export const getMe = createAsyncThunk(
  'auth/getMe',
  async (_, { rejectWithValue }) => {
    try {
      return await authService.getMe();
    } catch (err: any) {
      return rejectWithValue(err.response?.data?.error || 'Session expired');
    }
  }
);

export const register = createAsyncThunk(
  'auth/register',
  async (input: RegisterInput, { rejectWithValue }) => {
    try {
      return await authService.register(input);
    } catch (err: any) {
      return rejectWithValue(err.response?.data?.error || 'Registration failed');
    }
  }
);

export const login = createAsyncThunk(
  'auth/login',
  async (input: LoginInput, { rejectWithValue }) => {
    try {
      return await authService.login(input);
    } catch (err: any) {
      return rejectWithValue(err.response?.data?.error || 'Login failed');
    }
  }
);

export const logoutAsync = createAsyncThunk('auth/logout', async () => {
  try {
    await authService.logout();
  } catch {
    // clear local state even if server call fails
  }
});

// ── Initial State ─────────────────────────────────────────────────────────────

const initialState: AuthState = {
  user:            null,       // always null on load — getMe fills this in
  accessToken:     localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN),
  refreshToken:    localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN),
  isAuthenticated: false,      // false until server confirms the token is valid
  isLoading:       true,       // true so the app shows a loading screen on first paint
  error:           null,
};

// ── Slice ─────────────────────────────────────────────────────────────────────

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    setTokens(state, action: PayloadAction<{ accessToken: string; refreshToken: string }>) {
      state.accessToken  = action.payload.accessToken;
      state.refreshToken = action.payload.refreshToken;
      localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN,  action.payload.accessToken);
      localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, action.payload.refreshToken);
    },
    logout(state) {
      state.user            = null;
      state.accessToken     = null;
      state.refreshToken    = null;
      state.isAuthenticated = false;
      state.error           = null;
      clearStorage();
    },
    clearError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {

    // getMe — validates token on app load
    builder
      .addCase(getMe.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(getMe.fulfilled, (state, action) => {
        state.isLoading       = false;
        state.isAuthenticated = true;
        state.user            = action.payload;  // fresh user data from server
      })
      .addCase(getMe.rejected, (state) => {
        // Token was invalid or expired — clear everything
        state.isLoading       = false;
        state.isAuthenticated = false;
        state.user            = null;
        state.accessToken     = null;
        state.refreshToken    = null;
        clearStorage();
      });

    // Register
    builder
      .addCase(register.pending,   (state) => { state.isLoading = true; state.error = null; })
      .addCase(register.fulfilled, (state, action) => {
        state.isLoading       = false;
        state.isAuthenticated = true;
        state.user            = action.payload.user;
        state.accessToken     = action.payload.access_token;
        state.refreshToken    = action.payload.refresh_token;
        localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN,  action.payload.access_token);
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, action.payload.refresh_token);
      })
      .addCase(register.rejected, (state, action) => {
        state.isLoading = false;
        state.error     = action.payload as string;
      });

    // Login
    builder
      .addCase(login.pending,   (state) => { state.isLoading = true; state.error = null; })
      .addCase(login.fulfilled, (state, action) => {
        state.isLoading       = false;
        state.isAuthenticated = true;
        state.user            = action.payload.user;
        state.accessToken     = action.payload.access_token;
        state.refreshToken    = action.payload.refresh_token;
        localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN,  action.payload.access_token);
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, action.payload.refresh_token);
      })
      .addCase(login.rejected, (state, action) => {
        state.isLoading = false;
        state.error     = action.payload as string;
      });

    // Logout
    builder.addCase(logoutAsync.fulfilled, (state) => {
      state.user            = null;
      state.accessToken     = null;
      state.refreshToken    = null;
      state.isAuthenticated = false;
      clearStorage();
    });
  },
});

export const { setTokens, logout, clearError } = authSlice.actions;
export default authSlice.reducer;