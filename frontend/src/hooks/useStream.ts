import { useState, useCallback } from 'react';
import { streamService } from '../services/streamService';
import { 
    type StreamKeyResponse, 
    type StreamSettings 
} from '../types/stream.types';

export const useStream = () => {
  const [streamKey, setStreamKey] = useState<StreamKeyResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError]         = useState<string | null>(null);

  const fetchStreamKey = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await streamService.getStreamKey();
      setStreamKey(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load stream key');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const rotateKey = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await streamService.rotateStreamKey();
      setStreamKey(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to rotate key');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const updateSettings = useCallback(async (settings: StreamSettings) => {
    setIsLoading(true);
    setError(null);
    try {
      await streamService.updateSettings(settings);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update settings');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  return { streamKey, isLoading, error, fetchStreamKey, rotateKey, updateSettings };
};