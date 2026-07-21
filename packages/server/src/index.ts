import { serve } from '@hono/node-server';
import { createNodeWebSocket } from '@hono/node-ws';
import { app } from './app.js';
import { createStageWsHandler, startStageReaper } from './ws/stage-handler.js';

const { injectWebSocket, upgradeWebSocket } = createNodeWebSocket({ app });

// Mount stage WebSocket endpoint.
app.get('/ws/stage', upgradeWebSocket(createStageWsHandler()));

// Background reaper deletes stages idle for 10+ minutes (§10.2).
startStageReaper();

const port = Number(process.env.PORT || 3001);

const server = serve({ fetch: app.fetch, port }, (info) => {
  console.log(`[server] Listening on http://localhost:${info.port}`);
});

injectWebSocket(server);
