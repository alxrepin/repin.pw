export const TELEGRAM_BOT_UA = 'TelegramBot (like TwitterBot)';

// teletype copy
const SITE_VERIFICATION = 'g7j8/rPFXfhyrq5q0QQV7EsYWv4=';

const CHANNEL_CACHE_TTL_MS = 5 * 60 * 1000;

interface MediaItem {
  type: string;
  url: string;
  mime_type?: string;
}

interface PostData {
  title?: string;
  url?: string;
  text?: string;
  invert_media?: boolean;
  seo_title?: string;
  seo_description?: string;
  media?: MediaItem[];
  created_at: string;
}

interface ChannelData {
  name: string;
  title: string;
  avatar?: string;
}

function esc(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function stripTags(html: string): string {
  return html
    .replace(/<[^>]*>/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

function metaTag(property: string, content: string): string {
  return content ? `<meta property="${esc(property)}" content="${esc(content)}" />` : '';
}

function mediaHtml(media: MediaItem[], alt: string): string {
  return media
    .map(item => {
      if (item.type === 'photo') {
        return `<figure><img src="${esc(item.url)}" alt="${esc(alt)}" /></figure>`;
      }
      if (item.type === 'video') {
        return `<figure><video src="${esc(item.url)}"></video></figure>`;
      }
      return '';
    })
    .join('\n');
}

export function buildInstantViewHtml(
  post: PostData,
  channel: ChannelData,
  canonicalUrl: string,
): string {
  const title = post.title?.trim() || 'Без названия';
  const description = post.seo_description ?? stripTags(post.text ?? '').slice(0, 200);
  const cover = post.media?.find(item => item.type === 'photo');
  const ogImage = cover?.url ?? channel.avatar ?? '';
  const media = mediaHtml(post.media ?? [], title);

  const head = [
    metaTag('tg:site_verification', SITE_VERIFICATION),
    metaTag('telegram:channel', `@${channel.name}`),
    metaTag('article:published_time', post.created_at),
    metaTag('article:author', channel.title),
    metaTag('og:site_name', channel.title),
    metaTag('og:title', post.seo_title ?? title),
    metaTag('og:description', description),
    metaTag('og:image', ogImage),
    metaTag('og:url', canonicalUrl),
  ]
    .filter(Boolean)
    .join('\n');

  const body = [
    `<h1>${esc(title)}</h1>`,
    post.invert_media ? '' : media,
    post.text ?? '',
    post.invert_media ? media : '',
  ]
    .filter(Boolean)
    .join('\n');

  return `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="utf-8" />
<title>${esc(title)}</title>
${canonicalUrl ? `<link rel="canonical" href="${esc(canonicalUrl)}" />` : ''}
${head}
</head>
<body>
<div id="content" class="site-content article">
<article class="article__content">
<div class="entry-content">
${body}
</div>
</article>
</div>
</body>
</html>
`;
}

let channelCache: { data: ChannelData; expiresAt: number } | undefined;

async function fetchChannel(apiBase: string): Promise<ChannelData | null> {
  if (channelCache && channelCache.expiresAt > Date.now()) return channelCache.data;

  const response = await fetch(`${apiBase}/api/v1/channel`);
  if (!response.ok) return null;

  const payload = (await response.json()) as { data?: ChannelData };
  if (!payload.data) return null;

  channelCache = { data: payload.data, expiresAt: Date.now() + CHANNEL_CACHE_TTL_MS };
  return payload.data;
}

export async function renderInstantView(
  apiBase: string,
  siteUrl: string,
  slug: string,
): Promise<string | null> {
  try {
    const [postResponse, channel] = await Promise.all([
      fetch(`${apiBase}/api/v1/posts/${encodeURIComponent(slug)}`),
      fetchChannel(apiBase),
    ]);

    if (!postResponse.ok || !channel) return null;

    const payload = (await postResponse.json()) as { data?: PostData };
    if (!payload.data) return null;

    const canonicalSlug = payload.data.url ?? slug;
    const canonicalUrl = siteUrl ? `${siteUrl}/posts/${canonicalSlug}` : '';

    return buildInstantViewHtml(payload.data, channel, canonicalUrl);
  } catch (error) {
    console.error(`instant view render failed for ${slug}:`, error);
    return null;
  }
}
