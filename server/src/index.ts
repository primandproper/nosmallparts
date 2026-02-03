import express from 'express';
import cors from 'cors';
import { createServer } from 'http';
import { WebSocketServer, WebSocket } from 'ws';
import { ClientMessage, ServerMessage, Participant } from './shared/types';
import {
  createStage,
  getStageState,
  stageExists,
  addStageConnection,
  removeStageConnection,
  broadcastToStage,
  broadcastToCast,
  sendToParticipant,
  addParticipantToStage,
  toggleCastMember,
  setParticipantDirector,
  setParticipantReady,
  setParticipantAudience,
  selectPlayForStage,
  startStagePlay,
  advanceStageLine,
  resetStagePlay,
  getAvailablePlaysForStage,
  getAllScripts,
  getScriptById,
  createScript,
  updateScript,
  deleteScript,
  isUserScript,
} from './stages';
import { v4 as uuidv4 } from 'uuid';
import { generateToken, getLiveKitUrl, isLiveKitConfigured } from './livekit';

const app = express();
const server = createServer(app);
const wss = new WebSocketServer({ server });

app.use(cors());
app.use(express.json());

// REST endpoints
app.post('/api/stages', (req, res) => {
  const state = createStage();
  res.json({ stageId: state.id, state });
});

app.get('/api/stages/:stageId', (req, res) => {
  const state = getStageState(req.params.stageId);
  if (!state) {
    res.status(404).json({ error: 'Stage not found' });
    return;
  }
  res.json(state);
});

app.get('/api/stages/:stageId/exists', (req, res) => {
  res.json({ exists: stageExists(req.params.stageId) });
});

app.get('/api/plays', (req, res) => {
  res.json(getAvailablePlaysForStage());
});

// Scripts API endpoints
app.get('/api/scripts', (req, res) => {
  const scripts = getAllScripts();
  // Add metadata about whether each script is user-created
  const scriptsWithMeta = scripts.map((script) => ({
    ...script,
    isUserScript: isUserScript(script.id),
  }));
  res.json(scriptsWithMeta);
});

app.get('/api/scripts/:id', (req, res) => {
  const script = getScriptById(req.params.id);
  if (!script) {
    res.status(404).json({ error: 'Script not found' });
    return;
  }
  res.json({
    ...script,
    isUserScript: isUserScript(script.id),
  });
});

app.post('/api/scripts', (req, res) => {
  const { title, description, author, cast, lines } = req.body;

  if (!title || !cast || !lines) {
    res.status(400).json({ error: 'Missing required fields: title, cast, lines' });
    return;
  }

  const script = createScript({
    title,
    description: description || '',
    author: author || undefined,
    cast,
    lines,
  });

  res.status(201).json({
    ...script,
    isUserScript: true,
  });
});

app.put('/api/scripts/:id', (req, res) => {
  if (!isUserScript(req.params.id)) {
    res.status(403).json({ error: 'Cannot modify sample scripts' });
    return;
  }

  const updated = updateScript(req.params.id, req.body);
  if (!updated) {
    res.status(404).json({ error: 'Script not found' });
    return;
  }

  res.json({
    ...updated,
    isUserScript: true,
  });
});

app.delete('/api/scripts/:id', (req, res) => {
  if (!isUserScript(req.params.id)) {
    res.status(403).json({ error: 'Cannot delete sample scripts' });
    return;
  }

  if (deleteScript(req.params.id)) {
    res.status(204).send();
  } else {
    res.status(404).json({ error: 'Script not found' });
  }
});

app.post('/api/livekit/token', async (req, res) => {
  const { stageId, participantId, participantName } = req.body;

  if (!stageId || !participantId || !participantName) {
    res.status(400).json({ error: 'Missing required fields' });
    return;
  }

  if (!isLiveKitConfigured()) {
    res.status(503).json({ error: 'LiveKit not configured', url: '' });
    return;
  }

  try {
    const token = await generateToken(stageId, participantId, participantName);
    res.json({ token, url: getLiveKitUrl() });
  } catch (error) {
    res.status(500).json({ error: 'Failed to generate token' });
  }
});

// WebSocket handling
interface ConnectionContext {
  stageId?: string;
  participantId?: string;
}

wss.on('connection', (ws: WebSocket) => {
  const context: ConnectionContext = {};

  ws.on('message', (data: Buffer) => {
    try {
      const message: ClientMessage = JSON.parse(data.toString());
      handleMessage(ws, context, message);
    } catch (error) {
      console.error('Failed to parse message:', error);
    }
  });

  ws.on('close', () => {
    if (context.stageId && context.participantId) {
      removeStageConnection(context.stageId, context.participantId);
      const state = getStageState(context.stageId);
      if (state) {
        broadcastToStage(context.stageId, { type: 'stageState', state });
      }
    }
  });
});

function handleMessage(ws: WebSocket, context: ConnectionContext, message: ClientMessage): void {
  switch (message.type) {
    case 'join': {
      let stageId = message.stageId;
      console.log(`[Join] Request to join stage: "${stageId}" as "${message.participantName}"`);
      console.log(`[Join] Stage exists: ${stageExists(stageId)}`);

      // Create new stage if needed
      if (!stageId || !stageExists(stageId)) {
        console.log(`[Join] Creating new stage (requested: ${stageId})`);
        const newStage = createStage();
        stageId = newStage.id;
        console.log(`[Join] New stage created: ${stageId}`);
      }

      const participant = addParticipantToStage(stageId, message.participantName);
      if (!participant) {
        sendError(ws, 'Failed to join stage');
        return;
      }

      context.stageId = stageId;
      context.participantId = participant.id;

      addStageConnection(stageId, participant.id, ws);

      const state = getStageState(stageId);
      if (state) {
        // Send join confirmation to the new participant
        sendToParticipant(stageId, participant.id, {
          type: 'joined',
          participantId: participant.id,
          stageState: state,
        });

        // Send available plays
        sendToParticipant(stageId, participant.id, {
          type: 'availablePlays',
          plays: getAvailablePlaysForStage(),
        });

        // Broadcast updated state to all
        broadcastToStage(stageId, { type: 'stageState', state });
      }
      break;
    }

    case 'toggleCastMember': {
      if (!context.stageId) return;

      if (toggleCastMember(context.stageId, message.participantId, message.castMemberId)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, {
            type: 'castMembersUpdated',
            participants: state.participants,
          });
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }

    case 'setDirector': {
      if (!context.stageId) return;

      if (setParticipantDirector(context.stageId, message.participantId, message.isDirector)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }

    case 'setReady': {
      if (!context.stageId) return;

      if (setParticipantReady(context.stageId, message.participantId, message.isReady)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }

    case 'setAudience': {
      if (!context.stageId) return;

      if (setParticipantAudience(context.stageId, message.participantId, message.isAudience)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }

    case 'sendChat': {
      if (!context.stageId) return;

      const state = getStageState(context.stageId);
      if (!state) return;

      const sender = state.participants.find((p: Participant) => p.id === message.participantId);
      if (!sender) return;

      const chatMessage = {
        id: uuidv4(),
        senderId: sender.id,
        senderName: sender.name,
        text: message.text,
        castOnly: message.castOnly,
        timestamp: Date.now(),
      };

      // If cast-only, only send to non-audience participants
      if (message.castOnly) {
        broadcastToCast(context.stageId, { type: 'chatMessage', message: chatMessage });
      } else {
        broadcastToStage(context.stageId, { type: 'chatMessage', message: chatMessage });
      }
      break;
    }

    case 'selectPlay': {
      if (!context.stageId) return;

      if (selectPlayForStage(context.stageId, message.playId)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }

    case 'startPlay': {
      if (!context.stageId) return;

      if (startStagePlay(context.stageId)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, { type: 'countdownStart', duration: 3 });
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }

    case 'nextLine': {
      if (!context.stageId) return;

      if (advanceStageLine(context.stageId)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }

    case 'resetPlay': {
      if (!context.stageId) return;

      if (resetStagePlay(context.stageId)) {
        const state = getStageState(context.stageId);
        if (state) {
          broadcastToStage(context.stageId, { type: 'stageState', state });
        }
      }
      break;
    }
  }
}

function sendError(ws: WebSocket, errorMessage: string): void {
  const message: ServerMessage = { type: 'error', message: errorMessage };
  ws.send(JSON.stringify(message));
}

const PORT = process.env.PORT || 3001;
server.listen(Number(PORT), '0.0.0.0', () => {
  console.log(`Server running on http://0.0.0.0:${PORT}`);
  console.log(`LiveKit configured: ${isLiveKitConfigured()}`);
});
