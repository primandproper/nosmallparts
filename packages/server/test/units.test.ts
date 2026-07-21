import { describe, expect, it } from 'vitest';
import { parseChatMarkdown, truncateChatText, MAX_CHAT_LENGTH } from '@nsp/shared';
import { dedupeDisplayName, validateDisplayName } from '../src/display-name.js';
import { isAdminEmail } from '../src/admin.js';

describe('chat markdown (§6.2)', () => {
  it('renders bold, italic, code, strikethrough', () => {
    expect(parseChatMarkdown('**bold**')).toEqual([{ text: 'bold', bold: true }]);
    expect(parseChatMarkdown('*it*')).toEqual([{ text: 'it', italic: true }]);
    expect(parseChatMarkdown('_it_')).toEqual([{ text: 'it', italic: true }]);
    expect(parseChatMarkdown('`code`')).toEqual([{ text: 'code', code: true }]);
    expect(parseChatMarkdown('~~gone~~')).toEqual([{ text: 'gone', strike: true }]);
    expect(parseChatMarkdown('a **b** c')).toEqual([
      { text: 'a ' },
      { text: 'b', bold: true },
      { text: ' c' },
    ]);
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

describe('display names (§12.4)', () => {
  it('validates length, reserved tokens, profanity', () => {
    expect(validateDisplayName('Alice')).toEqual({ ok: true, name: 'Alice' });
    expect(validateDisplayName('  spaced   out  ')).toEqual({ ok: true, name: 'spaced out' });
    expect(validateDisplayName('x'.repeat(32)).ok).toBe(true);
    expect(validateDisplayName('x'.repeat(33)).ok).toBe(false);
    expect(validateDisplayName('').ok).toBe(false);
    for (const r of ['host', 'Admin', 'DIRECTOR', 'system', 'Server']) {
      expect(validateDisplayName(r).ok).toBe(false);
    }
    expect(validateDisplayName('fucknugget').ok).toBe(false);
  });

  it('dedupes with numeric suffixes within the length cap', () => {
    expect(dedupeDisplayName('Sam', [])).toBe('Sam');
    expect(dedupeDisplayName('Sam', ['sam'])).toBe('Sam (2)');
    expect(dedupeDisplayName('Sam', ['sam', 'sam (2)'])).toBe('Sam (3)');
    const long = 'x'.repeat(32);
    const deduped = dedupeDisplayName(long, [long]);
    expect(deduped.length).toBeLessThanOrEqual(32);
    expect(deduped.endsWith('(2)')).toBe(true);
  });
});

describe('admin emails (§2.4)', () => {
  it('matches case-insensitively against ADMIN_EMAILS', () => {
    // setup.ts sets ADMIN_EMAILS='admin@example.com, Boss@Example.COM'
    expect(isAdminEmail('admin@example.com')).toBe(true);
    expect(isAdminEmail('ADMIN@EXAMPLE.COM')).toBe(true);
    expect(isAdminEmail('boss@example.com')).toBe(true);
    expect(isAdminEmail('someone@example.com')).toBe(false);
    expect(isAdminEmail(null)).toBe(false);
    expect(isAdminEmail(undefined)).toBe(false);
  });
});
