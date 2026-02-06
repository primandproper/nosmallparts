import { createServer } from 'http';
import { parse } from 'url';

const dev = process.env.NODE_ENV !== 'production';
const hostname = process.env.HOSTNAME || '0.0.0.0';
const port = parseInt(process.env.PORT || '3000', 10);

// Initialize Next.js
// eslint-disable-next-line @typescript-eslint/no-require-imports
const next = require('next');
const app = next({ dev, hostname, port });
const handle = app.getRequestHandler();

// NOTE: WebSocket handling for stages has been moved to the Go backend.
// The frontend now connects directly to the Go backend at /ws/stages.
// Configure NEXT_PUBLIC_BACKEND_WS_HOST environment variable to set the backend host.

app.prepare().then(async () => {
  const server = createServer((req, res) => {
    const parsedUrl = parse(req.url!, true);
    handle(req, res, parsedUrl);
  });

  server.listen(port, hostname, () => {
    console.log(`> Ready on http://${hostname}:${port}`);
    console.log(
      `> Stages WebSocket handled by Go backend (NEXT_PUBLIC_BACKEND_WS_HOST: ${process.env.NEXT_PUBLIC_BACKEND_WS_HOST || 'localhost:8000'})`
    );
  });
});
