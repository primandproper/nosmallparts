import type { Metadata } from 'next';
import './globals.css';
import { AdSenseScript } from '@/components/AdSenseScript';

export const metadata: Metadata = {
  title: 'No Small Parts',
  description: 'A collaborative play performance platform',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        <AdSenseScript />
      </body>
    </html>
  );
}
