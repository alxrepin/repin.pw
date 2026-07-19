import { http } from '@/shared/api/http';
import type { ApiResponse } from '@/shared/types/api';
import type { Channel } from '../types';

interface ChannelData {
  id: number;
  name: string;
  title: string;
  description?: string;
  avatar?: string;
  subscriptions: number;
}

export async function fetchChannelRequest(): Promise<Channel> {
  const response = await http<ApiResponse<ChannelData, never>>('/api/v1/channel');
  const data = response.data;

  if (!data) throw new Error('channel not found');

  return {
    id: data.id,
    name: data.name,
    title: data.title,
    description: data.description,
    avatar: data.avatar,
    subscriptions: data.subscriptions,
    url: `https://t.me/${data.name}`,
  };
}
