export type PostMediaType =
  | 'photo'
  | 'video'
  | 'audio'
  | 'voice'
  | 'video_note'
  | 'gif'
  | 'document';

export interface PostMedia {
  id: number;
  type: PostMediaType;
  url: string;
  mimeType?: string;
  fileName?: string;
  size?: number;
  width?: number;
  height?: number;
  duration?: number;
}

export interface PostCover {
  url: string;
  width?: number;
  height?: number;
}

export interface PostSnippet {
  id: number;
  groupId: number;
  title: string;
  url: string;
  excerpt: string;
  cover?: PostCover;
  createdAt: string;
}

export interface Post {
  id: number;
  groupId: number;
  title: string;
  url: string;
  text: string;
  media: PostMedia[];
  /** Telegram "caption above media": render the gallery below the text. */
  invertMedia: boolean;
  prev?: PostSnippet;
  next?: PostSnippet;
  seoTitle?: string;
  seoDescription?: string;
  seoKeywords?: string;
  createdAt: string;
  updatedAt?: string;
}

export interface PostList {
  items: PostSnippet[];
  total: number;
}
