// Chat formatting rules.

/** Messages longer than this are truncated server-side before broadcast. */
export const MAX_CHAT_LENGTH = 256;

/**
 * Truncate a chat message to MAX_CHAT_LENGTH, appending an ellipsis when
 * anything was cut. Client input is not hard-capped; the server truncates too.
 */
export function truncateChatText(text: string): string {
  if (text.length <= MAX_CHAT_LENGTH) return text;
  return `${text.slice(0, MAX_CHAT_LENGTH - 1)}…`;
}

export interface ChatSegment {
  text: string;
  bold?: boolean;
  italic?: boolean;
  code?: boolean;
  strike?: boolean;
}

/**
 * Minimal, allowlist-only inline markdown for chat: bold (**x**), italic
 * (*x* / _x_), inline code (`x`), strikethrough (~~x~~). Everything else
 * (links, images, headings, blockquotes, tables, raw HTML, bare URLs) is left
 * as literal plain text, never rendered and never clickable.
 */
export function parseChatMarkdown(text: string): ChatSegment[] {
  const segments: ChatSegment[] = [];
  // Ordered by delimiter length so ** wins over *.
  const pattern = /(\*\*(.+?)\*\*|~~(.+?)~~|`([^`]+)`|\*([^*\s](?:[^*]*[^*\s])?)\*|_([^_\s](?:[^_]*[^_\s])?)_)/;
  let rest = text;
  while (rest.length > 0) {
    const m = rest.match(pattern);
    if (!m || m.index === undefined) {
      segments.push({ text: rest });
      break;
    }
    if (m.index > 0) segments.push({ text: rest.slice(0, m.index) });
    if (m[2] !== undefined) segments.push({ text: m[2], bold: true });
    else if (m[3] !== undefined) segments.push({ text: m[3], strike: true });
    else if (m[4] !== undefined) segments.push({ text: m[4], code: true });
    else if (m[5] !== undefined) segments.push({ text: m[5], italic: true });
    else if (m[6] !== undefined) segments.push({ text: m[6], italic: true });
    rest = rest.slice(m.index + m[0].length);
  }
  return segments.filter((s) => s.text.length > 0);
}
