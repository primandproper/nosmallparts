import { useEffect, useRef, useState } from 'react';
import { Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import { parseChatMarkdown } from '@nsp/shared';
import type { ChatMessage, SystemMessage } from '@nsp/shared';

interface Props {
  messages: ChatMessage[];
  systemMessages: SystemMessage[];
  isAudience: boolean;
  onSend: (text: string, castOnly: boolean) => void;
}

type Entry =
  | { kind: 'chat'; ts: number; msg: ChatMessage }
  | { kind: 'system'; ts: number; msg: SystemMessage };

/**
 * In-stage chat (§6). Messages support a small inline-markdown allowlist
 * (bold/italic/code/strikethrough); links, images, HTML, headings etc. render
 * as literal text. Input has no hard cap — the server truncates at 256.
 */
export function ChatPanel({ messages, systemMessages, isAudience, onSend }: Props) {
  const [text, setText] = useState('');
  const [castOnly, setCastOnly] = useState(false);
  const scrollRef = useRef<ScrollView | null>(null);

  const entries: Entry[] = [
    ...messages.map((m): Entry => ({ kind: 'chat', ts: m.timestamp, msg: m })),
    ...systemMessages.map((m): Entry => ({ kind: 'system', ts: m.timestamp, msg: m })),
  ].sort((a, b) => a.ts - b.ts);

  useEffect(() => {
    scrollRef.current?.scrollToEnd({ animated: false });
  }, [entries.length]);

  const submit = () => {
    const t = text.trim();
    if (!t) return;
    onSend(t, castOnly && !isAudience);
    setText('');
  };

  return (
    <View className="flex-1 min-h-0">
      <ScrollView
        ref={scrollRef}
        className="flex-1 min-h-0"
        contentContainerStyle={{ padding: 8, gap: 4 }}
      >
        {entries.length === 0 ? (
          <Text className="text-xs text-stage-muted">No messages yet. Say hi!</Text>
        ) : null}
        {entries.map((e) =>
          e.kind === 'system' ? (
            <Text key={e.msg.id} className="text-xs italic text-stage-muted">
              {e.msg.text}
            </Text>
          ) : (
            <View key={e.msg.id}>
              <Text className="text-xs">
                <Text
                  className={
                    e.msg.castOnly ? 'font-semibold text-amber-400' : 'font-semibold text-stage-gold'
                  }
                >
                  {e.msg.senderName}
                  {e.msg.castOnly ? ' (cast)' : ''}:{' '}
                </Text>
                <MarkdownText text={e.msg.text} />
              </Text>
            </View>
          )
        )}
      </ScrollView>
      <View className="border-t border-stage-border p-2 gap-2">
        {!isAudience ? (
          <Pressable
            onPress={() => setCastOnly((v) => !v)}
            accessibilityRole="button"
            className="self-start"
          >
            <Text className={`text-[10px] ${castOnly ? 'text-amber-400' : 'text-stage-muted'}`}>
              {castOnly ? '● Cast only' : '○ Everyone'}
            </Text>
          </Pressable>
        ) : null}
        <View className="flex-row gap-2">
          <TextInput
            value={text}
            onChangeText={setText}
            onSubmitEditing={submit}
            blurOnSubmit={false}
            placeholder="Message…"
            placeholderTextColor="#8a8374"
            className="flex-1 rounded-md border border-stage-border bg-stage-bg px-2 py-1.5 text-xs text-stage-text"
          />
          <Pressable
            onPress={submit}
            accessibilityRole="button"
            accessibilityLabel="Send message"
            className="rounded-md bg-stage-gold px-3 py-1.5"
          >
            <Text className="text-xs font-semibold text-stage-bg">Send</Text>
          </Pressable>
        </View>
      </View>
    </View>
  );
}

function MarkdownText({ text }: { text: string }) {
  const segments = parseChatMarkdown(text);
  return (
    <>
      {segments.map((s, i) => (
        <Text
          key={i}
          className={[
            'text-stage-text',
            s.bold ? 'font-bold' : '',
            s.italic ? 'italic' : '',
            s.code ? 'font-mono bg-stage-surface text-amber-200' : '',
          ]
            .filter(Boolean)
            .join(' ')}
          style={s.strike ? { textDecorationLine: 'line-through' } : undefined}
        >
          {s.text}
        </Text>
      ))}
    </>
  );
}
