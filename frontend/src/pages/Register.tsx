import { useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '../hooks/useAuth';
import { 
    type RegisterInput 
} from '../types/auth.types';
import styles from './Register.module.css';

const schema = z.object({
  email:    z.string().email('Enter a valid email address'),
  username: z.string()
    .min(3,  'Username must be at least 3 characters')
    .max(30, 'Username must be 30 characters or fewer')
    .regex(/^[a-zA-Z0-9_]+$/, 'Only letters, numbers and underscores'),
  password: z.string()
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Include at least one uppercase letter')
    .regex(/[0-9]/,  'Include at least one number'),
  confirmPassword: z.string(),
}).refine((d) => d.password === d.confirmPassword, {
  message: "Passwords don't match",
  path:    ['confirmPassword'],
});

type FormValues = z.infer<typeof schema>;

export default function Register() {
  const navigate = useNavigate();
  const { register: registerUser, isAuthenticated, isLoading, error, clearError } = useAuth();

  const { register, handleSubmit, watch, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
  });

  const password = watch('password', '');

  useEffect(() => { if (isAuthenticated) navigate('/dashboard'); }, [isAuthenticated, navigate]);
  useEffect(() => { return () => clearError(); }, []);

  const onSubmit = async (data: FormValues) => {
    const success = await registerUser({ email: data.email, username: data.username, password: data.password } as RegisterInput);
    if (success) navigate('/dashboard');
  };

  const getStrength = (pw: string) => {
    let score = 0;
    if (pw.length >= 8)          score++;
    if (/[A-Z]/.test(pw))        score++;
    if (/[0-9]/.test(pw))        score++;
    if (/[^A-Za-z0-9]/.test(pw)) score++;
    return score;
  };

  const strength = getStrength(password);
  const strengthColors = ['', '#ef4444', '#f97316', '#eab308', '#22c55e'];
  const strengthLabels = ['', 'Weak', 'Fair', 'Good', 'Strong'];

  return (
    <div className={styles.page}>
      <div className={styles.bgNoise} />

      <div className={styles.topBar}>
        <Link to="/" className={styles.logo}>
          <span className={styles.logoIcon}>◈</span>
          <span className={styles.logoText}>STREAMR</span>
        </Link>
        <div className={styles.topBarRight}>
          <span className={styles.topBarText}>Already have an account?</span>
          <Link to="/login" className={styles.topBarLink}>Sign in</Link>
        </div>
      </div>

      <div className={styles.content}>
        <div className={styles.formCard}>
          <div className={styles.formHeader}>
            <div className={styles.stepBadge}><span>01</span> Create account</div>
            <h1>Start streaming today</h1>
            <p>Join thousands of creators. It's completely free.</p>
          </div>

          {error && (
            <div className={styles.errorBanner} role="alert">
              <span className={styles.errorIcon}>!</span>
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} className={styles.form} noValidate>
            <div className={styles.row}>
              <div className={styles.field}>
                <label className={styles.label}>Email address</label>
                <input
                  type="email" autoComplete="email" placeholder="you@example.com"
                  className={`${styles.input} ${errors.email ? styles.inputError : ''}`}
                  {...register('email')}
                />
                {errors.email && <span className={styles.fieldError}>{errors.email.message}</span>}
              </div>

              <div className={styles.field}>
                <label className={styles.label}>Username</label>
                <div className={styles.inputWrapper}>
                  <span className={styles.inputPrefix}>@</span>
                  <input
                    type="text" autoComplete="username" placeholder="yourchannel"
                    className={`${styles.input} ${styles.inputPrefixed} ${errors.username ? styles.inputError : ''}`}
                    {...register('username')}
                  />
                </div>
                {errors.username && <span className={styles.fieldError}>{errors.username.message}</span>}
              </div>
            </div>

            <div className={styles.field}>
              <label className={styles.label}>Password</label>
              <input
                type="password" autoComplete="new-password" placeholder="••••••••"
                className={`${styles.input} ${errors.password ? styles.inputError : ''}`}
                {...register('password')}
              />
              {password.length > 0 && (
                <div className={styles.strengthMeter}>
                  <div className={styles.strengthBars}>
                    {[1, 2, 3, 4].map((n) => (
                      <div key={n} className={styles.strengthBar}
                        style={{ background: strength >= n ? strengthColors[strength] : '#1f2937', transition: 'background 0.3s' }}
                      />
                    ))}
                  </div>
                  <span className={styles.strengthLabel} style={{ color: strengthColors[strength] }}>
                    {strengthLabels[strength]}
                  </span>
                </div>
              )}
              {errors.password && <span className={styles.fieldError}>{errors.password.message}</span>}
            </div>

            <div className={styles.field}>
              <label className={styles.label}>Confirm password</label>
              <input
                type="password" autoComplete="new-password" placeholder="••••••••"
                className={`${styles.input} ${errors.confirmPassword ? styles.inputError : ''}`}
                {...register('confirmPassword')}
              />
              {errors.confirmPassword && <span className={styles.fieldError}>{errors.confirmPassword.message}</span>}
            </div>

            <p className={styles.terms}>
              By creating an account you agree to our{' '}
              <Link to="/terms" className={styles.termsLink}>Terms of Service</Link>
              {' '}and{' '}
              <Link to="/privacy" className={styles.termsLink}>Privacy Policy</Link>.
            </p>

            <button type="submit" className={styles.submitBtn} disabled={isLoading}>
              {isLoading ? (
                <span className={styles.spinner} />
              ) : (
                <>Create free account <span className={styles.btnArrow}>→</span></>
              )}
            </button>
          </form>
        </div>

        <div className={styles.featureList}>
          <h3>Everything you need to go live</h3>
          <ul>
            {[
              ['◈', 'OBS-compatible RTMP ingest'],
              ['◉', 'Live chat for your community'],
              ['◎', 'Subscriber notifications'],
              ['◆', 'Low-latency HLS delivery'],
            ].map(([icon, text]) => (
              <li key={text} className={styles.featureItem}>
                <span className={styles.featureIcon}>{icon}</span>
                <span>{text}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}