import { useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '../hooks/useAuth';
import { 
    type LoginInput 
} from '../types/auth.types';
import styles from './Login.module.css';

const schema = z.object({
  email:    z.string().email('Enter a valid email'),
  password: z.string().min(1, 'Password is required'),
});

type FormValues = z.infer<typeof schema>;

export default function Login() {
  const navigate = useNavigate();
  const { login, isAuthenticated, isLoading, error, clearError } = useAuth();

  const { register, handleSubmit, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
  });

  useEffect(() => {
    if (isAuthenticated) navigate('/dashboard');
  }, [isAuthenticated, navigate]);

  useEffect(() => { return () => clearError(); }, []);

  const onSubmit = async (data: FormValues) => {
    const success = await login(data as LoginInput);
    if (success) navigate('/dashboard');
  };

  return (
    <div className={styles.page}>
      <div className={styles.bgNoise} />

      <div className={styles.brandPanel}>
        <div className={styles.brandContent}>
          <div className={styles.logo}>
            <span className={styles.logoIcon}>◈</span>
            <span className={styles.logoText}>STREAMR</span>
          </div>
          <div className={styles.brandTagline}>
            <h1>Go live.<br />Get seen.</h1>
            <p>Stream to your audience in seconds. No setup required.</p>
          </div>
          <div className={styles.brandStats}>
            <div className={styles.stat}>
              <span className={styles.statNum}>12K+</span>
              <span className={styles.statLabel}>Active streamers</span>
            </div>
            <div className={styles.statDivider} />
            <div className={styles.stat}>
              <span className={styles.statNum}>98ms</span>
              <span className={styles.statLabel}>Avg. latency</span>
            </div>
          </div>
        </div>
        <div className={styles.gridLines} />
      </div>

      <div className={styles.formPanel}>
        <div className={styles.formCard}>
          <div className={styles.formHeader}>
            <h2>Welcome back</h2>
            <p>Sign in to your account to continue</p>
          </div>

          {error && (
            <div className={styles.errorBanner} role="alert">
              <span className={styles.errorIcon}>!</span>
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} className={styles.form} noValidate>
            <div className={styles.field}>
              <label htmlFor="email" className={styles.label}>Email address</label>
              <input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                className={`${styles.input} ${errors.email ? styles.inputError : ''}`}
                {...register('email')}
              />
              {errors.email && <span className={styles.fieldError}>{errors.email.message}</span>}
            </div>

            <div className={styles.field}>
              <div className={styles.labelRow}>
                <label htmlFor="password" className={styles.label}>Password</label>
                <Link to="/forgot-password" className={styles.forgotLink}>Forgot password?</Link>
              </div>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                className={`${styles.input} ${errors.password ? styles.inputError : ''}`}
                {...register('password')}
              />
              {errors.password && <span className={styles.fieldError}>{errors.password.message}</span>}
            </div>

            <button type="submit" className={styles.submitBtn} disabled={isLoading}>
              {isLoading ? (
                <span className={styles.spinner} />
              ) : (
                <>Sign in <span className={styles.btnArrow}>→</span></>
              )}
            </button>
          </form>

          <p className={styles.switchText}>
            Don't have an account?{' '}
            <Link to="/register" className={styles.switchLink}>Create one free</Link>
          </p>
        </div>
      </div>
    </div>
  );
}