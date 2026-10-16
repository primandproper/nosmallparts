// Ported from the TypeScript server's units.test.ts, each case under its
// original name.

import { describe, expect, it } from 'vitest';

import { MAX_CHAT_LENGTH, parseChatMarkdown, truncateChatText } from './chat';

describe('chat markdown (§6.2)', () => {
  it('renders bold, italic, code, strikethrough', () => {
    expect(parseChatMarkdown('**bold**')).toEqual([{ text: 'bold', bold: true }]);
    expect(parseChatMarkdown('*it*')).toEqual([{ text: 'it', italic: true }]);
    expect(parseChatMarkdown('_it_')).toEqual([{ text: 'it', italic: true }]);
    expect(parseChatMarkdown('`code`')).toEqual([{ text: 'code', code: true }]);
    expect(parseChatMarkdown('~~gone~~')).toEqual([{ text: 'gone', strike: true }]);
    expect(parseChatMarkdown('a **b** c')).toEqual([{ text: 'a ' }, { text: 'b', bold: true }, { text: ' c' }]);
  });

  it('leaves links, images, headings, html, and bare URLs as literal text', () => {
    const cases = [
      '[text](https://evil.example)',
      '![alt](https://evil.example/x.png)',
      '# heading',
      '> quote',
      '<script>alert(1)</script>',
      'https://example.com',
    ];
    for (const input of cases) {
      const segs = parseChatMarkdown(input);
      expect(segs.map((s) => s.text).join('')).toBe(input);
      expect(segs.every((s) => !s.bold && !s.italic && !s.code && !s.strike)).toBe(true);
    }
  });

  it('emojis pass through untouched', () => {
    expect(parseChatMarkdown('break a leg 🎭✨')).toEqual([{ text: 'break a leg 🎭✨' }]);
  });

  it('truncates at MAX_CHAT_LENGTH with ellipsis', () => {
    expect(MAX_CHAT_LENGTH).toBe(256);
    const long = 'x'.repeat(300);
    const cut = truncateChatText(long);
    expect(cut.length).toBe(256);
    expect(cut.endsWith('…')).toBe(true);
    expect(truncateChatText('short')).toBe('short');
    expect(truncateChatText('y'.repeat(256))).toBe('y'.repeat(256));
  });
});
