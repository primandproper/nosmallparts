'use client';

import { useEffect } from 'react';

interface AdBannerProps {
  slot: string;
  format?: 'auto' | 'horizontal' | 'vertical' | 'rectangle';
  responsive?: boolean;
  className?: string;
}

declare global {
  interface Window {
    adsbygoogle: unknown[];
  }
}

export function AdBanner({
  slot,
  format = 'auto',
  responsive = true,
  className = '',
}: AdBannerProps) {
  useEffect(() => {
    try {
      // Push ad when component mounts
      (window.adsbygoogle = window.adsbygoogle || []).push({});
    } catch (err) {
      console.error('AdSense error:', err);
    }
  }, []);

  return (
    <div className={`ad-container ${className}`}>
      <ins
        className="adsbygoogle"
        style={{ display: 'block' }}
        data-ad-client="ca-pub-8295255064000727"
        data-ad-slot={slot}
        data-ad-format={format}
        data-full-width-responsive={responsive ? 'true' : 'false'}
      />
    </div>
  );
}

// Placeholder component for development (shows where ads will appear)
export function AdBannerPlaceholder({ className = '' }: { className?: string }) {
  return (
    <div className={`ad-container ${className}`}>
      <div className="bg-[var(--charcoal)]/50 border border-dashed border-[var(--cream)]/20 rounded-xl p-4 text-center">
        <span className="text-[var(--cream)] opacity-30 text-sm">📢 Advertisement</span>
      </div>
    </div>
  );
}
