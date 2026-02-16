import type { Metadata } from 'next';
import Script from 'next/script';
import './globals.css';

export const metadata: Metadata = {
  title: 'No Small Parts',
  description: 'A collaborative play performance platform',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  // Use only NEXT_PUBLIC_* so server and client see the same value (prevents hydration mismatch).
  // ADSENSE_PUBLISHER_ID is server-only and would cause different <head> output.
  const adsenseClient = process.env.NEXT_PUBLIC_ADSENSE_PUBLISHER_ID || '';

  return (
    <html lang="en">
      <head>
        {adsenseClient && (
          <Script
            async
            src={`https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${adsenseClient}`}
            crossOrigin="anonymous"
            strategy="afterInteractive"
          />
        )}
      </head>
      <body>{children}</body>
    </html>
  );
}
