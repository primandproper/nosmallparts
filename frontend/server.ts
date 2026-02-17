import { createServer } from 'http';
import { parse } from 'url';
import { WebSocketServer } from 'ws';
import { handleStageWebSocket } from './lib/ws-stage-handler';

const dev = process.env.NODE_ENV !== 'production';
const hostname = process.env.HOSTNAME || '0.0.0.0';
const port = parseInt(process.env.PORT || '3000', 10);

// Initialize Next.js
// eslint-disable-next-line @typescript-eslint/no-require-imports
const next = require('next');
const app = next({ dev, hostname, port });
const handle = app.getRequestHandler();

app.prepare().then(async () => {
  const server = createServer((req, res) => {
    const parsedUrl = parse(req.url!, true);
    handle(req, res, parsedUrl);
  });

  const wss = new WebSocketServer({ noServer: true });

  server.on('upgrade', (req, socket, head) => {
    const { pathname } = parse(req.url ?? '', true);
    if (pathname === '/ws/stage') {
      wss.handleUpgrade(req, socket, head, (ws) => {
        wss.emit('connection', ws, req);
      });
    } else {
      socket.destroy();
    }
  });

  wss.on('connection', (ws, req) => {
    handleStageWebSocket(ws as import('ws').WebSocket, req);
  });

  server.listen(port, hostname, () => {
    console.log(`> Ready on http://${hostname}:${port}`);
    console.log(`> WebSocket stage at ws://${hostname}:${port}/ws/stage`);
  });
});
