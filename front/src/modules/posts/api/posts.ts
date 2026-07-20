import { type HttpError, http } from '@/shared/api/http';
import type { ApiResponse } from '@/shared/types/api';
import type { Post, PostCover, PostList, PostMedia, PostMediaType, PostSnippet } from '../types';

interface MediaItem {
  id: number;
  type: PostMediaType;
  url: string;
  mime_type?: string;
  file_name?: string;
  size?: number;
  width?: number;
  height?: number;
  duration?: number;
}

interface CoverItem {
  url: string;
  width?: number;
  height?: number;
}

interface ListItem {
  id: number;
  group_id: number;
  title?: string;
  url?: string;
  text?: string;
  cover?: CoverItem;
  created_at: string;
}

interface PostData extends ListItem {
  media?: MediaItem[];
  invert_media?: boolean;
  prev?: ListItem;
  next?: ListItem;
  seo_title?: string;
  seo_description?: string;
  seo_keywords?: string;
  updated_at?: string;
}

function stripTags(html: string): string {
  return html.replace(/<[^>]*>/g, '').trim();
}

function normalizeTitle(title: string): string {
  const trimmed = title.trim();
  const last = trimmed.slice(-1);

  if (last !== '.' && last !== ':') return trimmed;
  if (trimmed.endsWith('..')) return trimmed;

  return trimmed.slice(0, -1);
}

function toMedia(item: MediaItem): PostMedia {
  return {
    id: item.id,
    type: item.type,
    url: item.url,
    mimeType: item.mime_type,
    fileName: item.file_name,
    size: item.size,
    width: item.width,
    height: item.height,
    duration: item.duration,
  };
}

function toCover(item: CoverItem): PostCover {
  return { url: item.url, width: item.width, height: item.height };
}

function toSnippet(item: ListItem): PostSnippet {
  return {
    id: item.id,
    groupId: item.group_id,
    title: item.title ? normalizeTitle(item.title) : 'Без названия',
    url: item.url ?? String(item.id),
    excerpt: stripTags(item.text ?? ''),
    cover: item.cover ? toCover(item.cover) : undefined,
    createdAt: item.created_at,
  };
}

function toPost(data: PostData): Post {
  return {
    id: data.id,
    groupId: data.group_id,
    title: data.title ? normalizeTitle(data.title) : 'Без названия',
    url: data.url ?? String(data.id),
    text: data.text ?? '',
    media: data.media?.map(toMedia) ?? [],
    invertMedia: data.invert_media ?? false,
    prev: data.prev ? toSnippet(data.prev) : undefined,
    next: data.next ? toSnippet(data.next) : undefined,
    seoTitle: data.seo_title,
    seoDescription: data.seo_description,
    seoKeywords: data.seo_keywords,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export async function fetchPostsRequest(page = 1, limit = 9): Promise<PostList> {
  const response = await http<ApiResponse<never, ListItem>>('/api/v1/posts', {
    query: { page, limit },
  });

  return {
    items: response.items.map(toSnippet),
    total: response.paginate?.total ?? 0,
  };
}

export async function fetchPostRequest(slug: string): Promise<Post | null> {
  try {
    const response = await http<ApiResponse<PostData, never>>(`/api/v1/posts/${slug}`);
    return response.data ? toPost(response.data) : null;
  } catch (error) {
    if ((error as HttpError).status === 404) return null;
    throw error;
  }
}
