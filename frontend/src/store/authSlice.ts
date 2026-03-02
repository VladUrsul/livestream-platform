import { createSlice, 
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

const initialState: AuthState = {
  user:            null,
  accessToken:     localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN),
  refreshToken:    localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN),
  isAuthenticated: !!localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN),
  isLoading:       false,
  error:           null,
};

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
      localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN);
      localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
    },
    clearError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(register.pending,   (state) => { state.isLoading = true;  state.error = null; })
      .addCase(register.fulfilled, (state, action) => {
        state.isLoading       = false;
        state.isAuthenticated = true;
        state.user            = action.payload.user;
        state.accessToken     = action.payload.access_token;
        state.refreshToken    = action.payload.refresh_token;
        localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN,  action.payload.access_token);
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, action.payload.refresh_token);
      })
      .addCase(register.rejected,  (state, action) => { state.isLoading = false; state.error = action.payload as string; });

    builder
      .addCase(login.pending,   (state) => { state.isLoading = true;  state.error = null; })
      .addCase(login.fulfilled, (state, action) => {
        state.isLoading       = false;
        state.isAuthenticated = true;
        state.user            = action.payload.user;
        state.accessToken     = action.payload.access_token;
        state.refreshToken    = action.payload.refresh_token;
        localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN,  action.payload.access_token);
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, action.payload.refresh_token);
      })
      .addCase(login.rejected,  (state, action) => { state.isLoading = false; state.error = action.payload as string; });

    builder.addCase(logoutAsync.fulfilled, (state) => {
      state.user            = null;
      state.accessToken     = null;
      state.refreshToken    = null;
      state.isAuthenticated = false;
      localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN);
      localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
    });
  },
});

export const { setTokens, logout, clearError } = authSlice.actions;
export default authSlice.reducer;