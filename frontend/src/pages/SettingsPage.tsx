import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useAuth } from '../hooks/useAuth';
import { userService } from '../services/userService';
import styles from './SettingsPage.module.css';

const schema = z.object({
  display_name: z.string().min(1, 'Display name is required').max(60, 'Max 60 characters'),
  bio: z.string().max(500, 'Max 500 characters').default(''),
  avatar_url: z.string().url('Enter a valid URL').or(z.literal('')).default(''),
});

type FormValues = z.infer<typeof schema>;

export default function SettingsPage() {
  const { user } = useAuth();

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { register, handleSubmit, formState: { errors }, reset } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      display_name: '',
      bio: '',
      avatar_url: '',
    },
  });

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      try {
        setError(null);
        const me = await userService.getMe();
        if (cancelled) return;
        reset({
          display_name: me.display_name ?? '',
          bio: me.bio ?? '',
          avatar_url: me.avatar_url ?? '',
        });
      } catch (e: any) {
        if (cancelled) return;
        setError(e?.response?.data?.error || 'Failed to load profile');
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    // `user` is available after auth; still rely on getMe() for fresh profile data.
    if (user) load();
    return () => {
      cancelled = true;
    };
  }, [user, reset]);

  const onSubmit = async (data: FormValues) => {
    if (!data.display_name) return;
    setSaving(true);
    setError(null);
    try {
      await userService.updateProfile({
        display_name: data.display_name.trim(),
        bio: data.bio.trim() ? data.bio.trim() : undefined,
        avatar_url: data.avatar_url.trim() ? data.avatar_url.trim() : undefined,
      });
    } catch (e: any) {
      setError(e?.response?.data?.error || 'Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.centerState}>
          <span className={styles.spinner} />
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Settings</h1>
          <p className={styles.subtitle}>Update your public profile</p>
        </div>
      </div>

      <form className={styles.card} onSubmit={handleSubmit(onSubmit)}>
        <div className={styles.formRow}>
          <label className={styles.label}>Display name</label>
          <input className={styles.input} {...register('display_name')} placeholder="Your name" />
          {errors.display_name && <span className={styles.errorText}>{errors.display_name.message}</span>}
        </div>

        <div className={styles.formRow}>
          <label className={styles.label}>Bio</label>
          <textarea className={styles.textarea} rows={4} {...register('bio')} placeholder="Tell people about your stream..." />
          {errors.bio && <span className={styles.errorText}>{errors.bio.message}</span>}
        </div>

        <div className={styles.formRow}>
          <label className={styles.label}>Avatar URL</label>
          <input className={styles.input} {...register('avatar_url')} placeholder="https://..." />
          {errors.avatar_url && <span className={styles.errorText}>{errors.avatar_url.message}</span>}
        </div>

        {error && (
          <div className={styles.errorBanner} role="alert">
            {error}
          </div>
        )}

        <div className={styles.actions}>
          <button className={styles.saveBtn} type="submit" disabled={saving}>
            {saving ? (
              <>
                <span className={styles.spinnerSmall} /> Saving...
              </>
            ) : (
              'Save changes'
            )}
          </button>
        </div>
      </form>
    </div>
  );
}

