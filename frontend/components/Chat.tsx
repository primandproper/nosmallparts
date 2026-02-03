'use client';

import { useState, useRef, useEffect } from 'react';
import { ChatMessage, Participant } from '@/lib/types';

interface ChatProps {
  messages: ChatMessage[];
  currentParticipant: Participant | null;
  onSendMessage: (text: string, castOnly: boolean) => void;
}

export function Chat({ messages, currentParticipant, onSendMessage }: ChatProps) {
  const [inputText, setInputText] = useState('');
  const [castOnly, setCastOnly] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const prevMessageCountRef = useRef(0);

  const isAudience = currentParticipant?.isAudience || false;

  // Auto-scroll to bottom when NEW messages arrive (not on initial mount)
  useEffect(() => {
    // Only scroll if we have new messages (not on initial render)
    if (messages.length > prevMessageCountRef.current && messagesContainerRef.current) {
      // Scroll within the container only, not the whole page
      messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight;
    }
    prevMessageCountRef.current = messages.length;
  }, [messages]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputText.trim()) return;

    onSendMessage(inputText.trim(), castOnly);
    setInputText('');
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="card-elegant rounded-2xl overflow-hidden flex flex-col h-[400px]">
      {/* Header */}
      <div className="p-3 border-b border-[var(--cream)]/10 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-lg">💬</span>
          <h3 className="font-[Cormorant_Garamond] text-lg font-semibold text-[var(--gold)]">
            Stage Chat
          </h3>
        </div>
        <div className="text-xs text-[var(--cream)] opacity-40">
          {messages.length} message{messages.length !== 1 ? 's' : ''}
        </div>
      </div>

      {/* Messages */}
      <div ref={messagesContainerRef} className="flex-1 overflow-y-auto p-3 space-y-3">
        {messages.length === 0 ? (
          <div className="h-full flex items-center justify-center">
            <p className="text-[var(--cream)] opacity-30 text-sm text-center">
              No messages yet.
              <br />
              Start the conversation!
            </p>
          </div>
        ) : (
          messages.map((msg) => {
            const isMe = msg.senderId === currentParticipant?.id;

            return (
              <div key={msg.id} className={`flex flex-col ${isMe ? 'items-end' : 'items-start'}`}>
                {/* Sender name & time */}
                <div
                  className={`flex items-center gap-2 text-xs mb-1 ${isMe ? 'flex-row-reverse' : ''}`}
                >
                  <span className="text-[var(--cream)] opacity-60">{msg.senderName}</span>
                  <span className="text-[var(--cream)] opacity-30">
                    {formatTime(msg.timestamp)}
                  </span>
                  {msg.castOnly && (
                    <span className="px-1.5 py-0.5 rounded bg-[var(--gold)]/20 text-[var(--gold)] text-[10px] uppercase tracking-wide">
                      Cast
                    </span>
                  )}
                </div>

                {/* Message bubble */}
                <div
                  className={`max-w-[80%] px-3 py-2 rounded-2xl text-sm ${
                    isMe
                      ? 'bg-[var(--gold)]/20 text-[var(--cream)] rounded-br-md'
                      : msg.castOnly
                        ? 'bg-[var(--gold)]/10 border border-[var(--gold)]/30 text-[var(--cream)] rounded-bl-md'
                        : 'bg-[var(--charcoal)] text-[var(--cream)] rounded-bl-md'
                  }`}
                >
                  {msg.text}
                </div>
              </div>
            );
          })
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="p-3 border-t border-[var(--cream)]/10">
        {/* Cast-only toggle - only for non-audience */}
        {!isAudience && (
          <div className="flex items-center gap-2 mb-2">
            <button
              type="button"
              onClick={() => setCastOnly(!castOnly)}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                castOnly
                  ? 'bg-[var(--gold)]/20 text-[var(--gold)] border border-[var(--gold)]/30'
                  : 'bg-[var(--charcoal)] text-[var(--cream)] opacity-60 hover:opacity-100'
              }`}
            >
              <span>{castOnly ? '🎭' : '🌐'}</span>
              <span>{castOnly ? 'Cast Only' : 'Everyone'}</span>
            </button>
            {castOnly && (
              <span className="text-[var(--cream)] opacity-40 text-xs">
                Only cast members will see this
              </span>
            )}
          </div>
        )}

        <div className="flex gap-2">
          <input
            type="text"
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder={
              isAudience
                ? 'Send a message...'
                : castOnly
                  ? 'Message to cast...'
                  : 'Send a message...'
            }
            className="flex-1 input-elegant px-4 py-2 rounded-xl text-sm"
          />
          <button
            type="submit"
            disabled={!inputText.trim()}
            className="btn-primary px-4 py-2 rounded-xl text-sm disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send
          </button>
        </div>
      </form>
    </div>
  );
}
