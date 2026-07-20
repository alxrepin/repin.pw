import type { Channel } from '@/modules/channel/types';

export function channelDescription(channel: Channel): string {
  return channel.description || `Зеркало Telegram-канала @${channel.name}`;
}

export function jsonLd(data: object): string {
  return JSON.stringify(data).replace(/</g, '\\u003C');
}
