'use client';

import { useEffect } from 'react';

interface AdBannerProps {
  /** Publisher ID (ca-pub-...) – from AdBannerServer or env */
  client: string;
  /** Ad slot ID from AdSense dashboard */
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
  client,
  slot,
  format = 'auto',
  responsive = true,
  className = '',
}: AdBannerProps) {
  useEffect(() => {
    const isDebug =
      typeof window !== 'undefined' &&
      new URLSearchParams(window.location.search).get('ads_debug') === '1';
    if (isDebug && (client || slot)) {
      console.log('[AdSense] Ad slot diagnostic:', {
        client: client ? `ca-pub-***${client.slice(-4)}` : 'missing',
        slot: slot || 'missing',
        adsbygoogleLoaded: typeof window.adsbygoogle !== 'undefined',
      });
    }
    try {
      (window.adsbygoogle = window.adsbygoogle || []).push({});
    } catch (err) {
      console.error('AdSense error:', err);
    }
  }, [client, slot]);

  if (!client || !slot) return null;

  return (
    <div className={`ad-container ${className}`}>
      <ins
        className="adsbygoogle"
        style={{ display: 'block' }}
        data-ad-client={client}
        data-ad-slot={slot}
        data-ad-format={format}
        data-full-width-responsive={responsive ? 'true' : 'false'}
      />
    </div>
  );
}
