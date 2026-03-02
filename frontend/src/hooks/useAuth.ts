import { useDispatch, useSelector } from 'react-redux';
import { 
    type AppDispatch, 
    type RootState 
} from '../store';
import { register, login, logoutAsync, clearError } from '../store/authSlice';
import { 
    type LoginInput, 
    type RegisterInput 
} from '../types/auth.types';

export const useAuth = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { user, isAuthenticated, isLoading, error } = useSelector(
    (state: RootState) => state.auth
  );

  const handleRegister = async (input: RegisterInput) => {
    const result = await dispatch(register(input));
    return !result.type.endsWith('rejected');
  };

  const handleLogin = async (input: LoginInput) => {
    const result = await dispatch(login(input));
    return !result.type.endsWith('rejected');
  };

  const handleLogout = () => {
    dispatch(logoutAsync());
  };

  const handleClearError = () => {
    dispatch(clearError());
  };

  return {
    user,
    isAuthenticated,
    isLoading,
    error,
    register: handleRegister,
    login:    handleLogin,
    logout:   handleLogout,
    clearError: handleClearError,
  };
};