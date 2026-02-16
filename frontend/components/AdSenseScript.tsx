'use client';

import { useEffect } from 'react';

/**
 * Loads the AdSense script after hydration. Must be a client component to avoid
 * hydration mismatch - any conditional script in the layout causes different
 * server vs client <head> output and breaks React hydration.
 *
 * Diagnosis: Add ?ads_debug=1 to any URL and check the console for AdSense config status.
 */
export function AdSenseScript() {
  useEffect(() => {
    const client = process.env.NEXT_PUBLIC_ADSENSE_PUBLISHER_ID;
    const isDebug =
      typeof window !== 'undefined' && new URLSearchParams(window.location.search).get('ads_debug') === '1';

    if (isDebug) {
      console.log('[AdSense] Diagnostic:', {
        'NEXT_PUBLIC_ADSENSE_PUBLISHER_ID': client
          ? `ca-pub-***${client.slice(-4)}`
          : 'NOT SET (empty at build time)',
        scriptWillLoad: !!client,
        hint: !client
          ? 'NEXT_PUBLIC_* must be passed as --build-arg during docker build. Runtime env vars are not available client-side.'
          : undefined,
      });
    }

    if (!client) return;

    const script = document.createElement('script');
    script.async = true;
    script.src = `https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${client}`;
    script.crossOrigin = 'anonymous';
    document.head.appendChild(script);
  }, []);

  return null;
}
