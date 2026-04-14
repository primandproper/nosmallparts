import { serve } from '@hono/node-server';
import { createNodeWebSocket } from '@hono/node-ws';
import { app } from './app.js';
import { createStageWsHandler } from './ws/stage-handler.js';

const { injectWebSocket, upgradeWebSocket } = createNodeWebSocket({ app });

// Mount stage WebSocket endpoint.
app.get('/ws/stage', upgradeWebSocket(createStageWsHandler()));

const port = Number(process.env.PORT || 3001);

const server = serve({ fetch: app.fetch, port }, (info) => {
  console.log(`[server] Listening on http://localhost:${info.port}`);
});

injectWebSocket(server);
