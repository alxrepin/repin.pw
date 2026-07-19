import { useAsyncData } from '@/shared/composables/useAsyncData';
import { fetchChannelRequest } from '../api/channel';
import type { Channel } from '../types';

export function useChannel(): Promise<Channel> {
  return useAsyncData('channel', fetchChannelRequest);
}
