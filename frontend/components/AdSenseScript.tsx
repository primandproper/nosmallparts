'use client';

import { useEffect } from 'react';

/**
 * Loads the AdSense script after hydration. Must be a client component to avoid
 * hydration mismatch - any conditional script in the layout causes different
 * server vs client <head> output and breaks React hydration.
 */
export function AdSenseScript() {
  useEffect(() => {
    const client = process.env.NEXT_PUBLIC_ADSENSE_PUBLISHER_ID;
    if (!client) return;

    const script = document.createElement('script');
    script.async = true;
    script.src = `https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${client}`;
    script.crossOrigin = 'anonymous';
    document.head.appendChild(script);
  }, []);

  return null;
}
