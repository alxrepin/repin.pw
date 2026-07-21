import { describe, expect, test } from 'bun:test';
import { buildInstantViewHtml } from './instantview';

const channel = { name: 'repin', title: 'Канал Репина', avatar: 'https://cdn/avatar.jpg' };

describe('buildInstantViewHtml', () => {
  test('escapes html in title and meta content', () => {
    const html = buildInstantViewHtml(
      {
        title: '<script>alert(1)</script> & "кавычки"',
        seo_description: 'desc with "quotes" & <tags>',
        created_at: '2026-01-01T00:00:00Z',
      },
      channel,
      'https://site/posts/1',
    );

    expect(html).not.toContain('<script>alert(1)</script>');
    expect(html).toContain('&lt;script&gt;alert(1)&lt;/script&gt; &amp; &quot;кавычки&quot;');
    expect(html).toContain('content="desc with &quot;quotes&quot; &amp; &lt;tags&gt;"');
  });

  test('renders photo and video media', () => {
    const html = buildInstantViewHtml(
      {
        title: 'Пост',
        media: [
          { type: 'photo', url: 'https://cdn/p.jpg' },
          { type: 'video', url: 'https://cdn/v.mp4', width: 640 },
          { type: 'audio', url: 'https://cdn/a.mp3' },
        ],
        created_at: '2026-01-01T00:00:00Z',
      },
      channel,
      '',
    );

    expect(html).toContain('<img src="https://cdn/p.jpg"');
    expect(html).toContain('<video class="embed-viedo" src="https://cdn/v.mp4"');
    expect(html).toContain('width="640"');
    expect(html).not.toContain('a.mp3'); // unsupported media types are dropped
  });

  test('invert_media puts media after the text', () => {
    const post = {
      title: 'Пост',
      text: '<p>тело</p>',
      media: [{ type: 'photo', url: 'https://cdn/p.jpg' }],
      created_at: '2026-01-01T00:00:00Z',
    };

    // og:image in <head> also carries the photo url, so compare <figure> positions
    const normal = buildInstantViewHtml(post, channel, '');
    expect(normal.indexOf('<figure>')).toBeLessThan(normal.indexOf('<p>тело</p>'));

    const inverted = buildInstantViewHtml({ ...post, invert_media: true }, channel, '');
    expect(inverted.indexOf('<figure>')).toBeGreaterThan(inverted.indexOf('<p>тело</p>'));
  });

  test('falls back to stripped text for description and avatar for og:image', () => {
    const html = buildInstantViewHtml(
      {
        title: 'Пост',
        text: '<p>первый абзац</p><p>второй</p>',
        created_at: '2026-01-01T00:00:00Z',
      },
      channel,
      '',
    );

    expect(html).toContain('content="первый абзац второй"');
    expect(html).toContain('content="https://cdn/avatar.jpg"'); // no photo → channel avatar
    expect(html).not.toContain('rel="canonical"'); // no canonical without a site url
  });

  test('prefers post photo over avatar and emits canonical', () => {
    const html = buildInstantViewHtml(
      {
        title: 'Пост',
        media: [{ type: 'photo', url: 'https://cdn/cover.jpg' }],
        created_at: '2026-01-01T00:00:00Z',
      },
      channel,
      'https://site/posts/42-slug',
    );

    expect(html).toContain('<meta property="og:image" content="https://cdn/cover.jpg"');
    expect(html).toContain('<link rel="canonical" href="https://site/posts/42-slug"');
  });
});
