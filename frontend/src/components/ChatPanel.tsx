import { type FormEvent, useEffect, useRef, useState } from 'react';

import type { ChatMessage, SystemMessage } from '../generated/nosmallparts/stage/v1/stage';
import { parseChatMarkdown } from '../lib/chat';

interface Props {
  messages: ChatMessage[];
  systemMessages: SystemMessage[];
  isAudience: boolean;
  onSend: (text: string, castOnly: boolean) => void;
}

type Entry = { kind: 'chat'; ts: number; msg: ChatMessage } | { kind: 'system'; ts: number; msg: SystemMessage };

const timeOf = (d: Date | undefined) => d?.getTime() ?? 0;

/**
 * In-stage chat (§6). Messages support a small inline-markdown allowlist
 * (bold, italic, code, strikethrough); links, images, HTML and headings render
 * as literal text. Input has no hard cap; the server truncates at 256.
 */
export function ChatPanel({ messages, systemMessages, isAudience, onSend }: Props) {
  const [text, setText] = useState('');
  const [castOnly, setCastOnly] = useState(false);
  const endRef = useRef<HTMLDivElement | null>(null);

  const entries: Entry[] = [
    ...messages.map((m): Entry => ({ kind: 'chat', ts: timeOf(m.timestamp), msg: m })),
    ...systemMessages.map((m): Entry => ({ kind: 'system', ts: timeOf(m.timestamp), msg: m })),
  ].sort((a, b) => a.ts - b.ts);

  useEffect(() => {
    endRef.current?.scrollIntoView({ block: 'end' });
  }, [entries.length]);

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const t = text.trim();
    if (!t) return;
    onSend(t, castOnly && !isAudience);
    setText('');
  };

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-2">
        {entries.length === 0 ? <p className="text-xs text-stage-muted">No messages yet. Say hi!</p> : null}
        {entries.map((e) =>
          e.kind === 'system' ? (
            <p key={e.msg.id} className="text-xs italic text-stage-muted">
              {e.msg.text}
            </p>
          ) : (
            <p key={e.msg.id} className="text-xs">
              <span className={e.msg.castOnly ? 'font-semibold text-amber-400' : 'font-semibold text-stage-gold'}>
                {e.msg.senderName}
                {e.msg.castOnly ? ' (cast)' : ''}:{' '}
              </span>
              <MarkdownText text={e.msg.text} />
            </p>
          ),
        )}
        <div ref={endRef} />
      </div>
      <form onSubmit={submit} className="flex flex-col gap-2 border-t border-stage-border p-2">
        {!isAudience ? (
          <button
            type="button"
            onClick={() => setCastOnly((v) => !v)}
            aria-pressed={castOnly}
            className={`self-start text-[10px] ${castOnly ? 'text-amber-400' : 'text-stage-muted'}`}
          >
            {castOnly ? '● Cast only' : '○ Everyone'}
          </button>
        ) : null}
        <div className="flex gap-2">
          <input
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Message…"
            className="min-w-0 flex-1 rounded-md border border-stage-border bg-stage-bg px-2 py-1.5 text-xs text-stage-text placeholder:text-stage-muted"
          />
          <button
            type="submit"
            aria-label="Send message"
            className="rounded-md bg-stage-gold px-3 py-1.5 text-xs font-semibold text-stage-bg"
          >
            Send
          </button>
        </div>
      </form>
    </div>
  );
}

function MarkdownText({ text }: { text: string }) {
  return (
    <>
      {parseChatMarkdown(text).map((s, i) => (
        <span
          key={i}
          className={[
            'text-stage-text',
            s.bold ? 'font-bold' : '',
            s.italic ? 'italic' : '',
            s.code ? 'bg-stage-surface font-mono text-amber-200' : '',
            s.strike ? 'line-through' : '',
          ]
            .filter(Boolean)
            .join(' ')}
        >
          {s.text}
        </span>
      ))}
    </>
  );
}
