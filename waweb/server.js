const express = require("express");
const cors = require("cors");
const QRCode = require("qrcode");
const fs = require("fs");
const path = require("path");
const { Client, LocalAuth } = require("whatsapp-web.js");

const PORT = Number(process.env.PORT || 3000);
const DATA_DIR = process.env.WA_WEB_AUTH_DIR || path.join(__dirname, ".data");
const AUTH_DIR = path.join(DATA_DIR, ".wwebjs_auth");
const CLIENT_ID = process.env.WA_WEB_CLIENT_ID || "dentdesk-session";

fs.mkdirSync(AUTH_DIR, { recursive: true });

const app = express();
app.use(cors());
app.use(express.json({ limit: "1mb" }));

let client = null;
let isReady = false;
let isAuthenticated = false;
let isInitializing = false;
let latestQrCode = null;
let latestInitError = null;
let currentClientState = "starting";
let loadingPercent = null;
let lastDisconnectReason = null;

const eventClients = new Set();
const recentlyBroadcastMessageIds = new Set();

function sendEvent(res, event, data) {
  res.write(`event: ${event}\n`);
  res.write(`data: ${JSON.stringify(data)}\n\n`);
}

function broadcast(event, data) {
  for (const res of eventClients) {
    try {
      sendEvent(res, event, data);
    } catch {
      eventClients.delete(res);
    }
  }
}

function statusPayload() {
  return {
    ready: isReady,
    authenticated: isAuthenticated,
    initializing: isInitializing,
    state: currentClientState,
    loadingPercent,
    hasQr: Boolean(latestQrCode),
    error: latestInitError,
    lastDisconnectReason,
  };
}

function broadcastStatus() {
  broadcast("status", statusPayload());
}

function normalizeLimit(value, fallback, max) {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) return fallback;
  return Math.min(parsed, max);
}

function getMessageId(message) {
  return message?.id?._serialized || message?.id?.id || "";
}

function getChatIdFromMessage(message) {
  if (!message) return "";
  if (message.fromMe && message.to) return message.to;
  return message.from || message.to || "";
}

function readableBody(message) {
  if (!message) return "";
  if (message.body) return message.body;
  if (message.hasMedia) return `[${message.type || "media"}]`;
  return "";
}

function serializeMessage(message) {
  const timestamp = Number(message?.timestamp || 0);
  return {
    id: getMessageId(message) || `${getChatIdFromMessage(message)}-${timestamp}-${Math.random()}`,
    chatId: getChatIdFromMessage(message),
    from: message?.from || "",
    to: message?.to || "",
    author: message?.author || null,
    body: readableBody(message),
    type: message?.type || "message",
    timestamp,
    createdAt: timestamp ? new Date(timestamp * 1000).toISOString() : new Date().toISOString(),
    fromMe: Boolean(message?.fromMe),
    hasMedia: Boolean(message?.hasMedia),
    ack: message?.ack ?? null,
  };
}

function getChatTimestamp(chat) {
  return Number(chat?.lastMessage?.timestamp || chat?.timestamp || 0);
}

function isNormalChat(chat) {
  const id = chat?.id?._serialized || "";
  if (!id) return false;
  if (id === "status@broadcast") return false;
  if (id.endsWith("@newsletter")) return false;
  if (id.endsWith("@broadcast")) return false;
  return true;
}

function serializeChat(chat) {
  const id = chat?.id?._serialized || "";
  return {
    id,
    name: chat?.name || chat?.pushname || chat?.formattedTitle || id.replace("@c.us", ""),
    isGroup: Boolean(chat?.isGroup),
    unreadCount: Number(chat?.unreadCount || 0),
    lastActivityTimestamp: getChatTimestamp(chat),
    lastMessage: chat?.lastMessage ? serializeMessage(chat.lastMessage) : null,
  };
}

function ensureReady(res) {
  if (!client || !isReady) {
    res.status(409).json({
      error: "WhatsApp Web is not connected yet. Scan the QR code first.",
      status: statusPayload(),
    });
    return false;
  }
  return true;
}

function broadcastMessage(message) {
  const id = getMessageId(message);
  if (id && recentlyBroadcastMessageIds.has(id)) return;
  if (id) {
    recentlyBroadcastMessageIds.add(id);
    setTimeout(() => recentlyBroadcastMessageIds.delete(id), 30000).unref?.();
  }

  const serialized = serializeMessage(message);
  broadcast("message", { chatId: serialized.chatId, message: serialized });
  broadcast("chats", { reason: "message", chatId: serialized.chatId });
}

function createClient() {
  return new Client({
    authStrategy: new LocalAuth({
      clientId: CLIENT_ID,
      dataPath: AUTH_DIR,
    }),
    takeoverOnConflict: true,
    puppeteer: {
      headless: true,
      executablePath: process.env.PUPPETEER_EXECUTABLE_PATH || undefined,
      args: [
        "--no-sandbox",
        "--disable-setuid-sandbox",
        "--disable-dev-shm-usage",
        "--disable-background-timer-throttling",
        "--disable-backgrounding-occluded-windows",
        "--disable-renderer-backgrounding",
      ],
    },
  });
}

async function initializeClient() {
  if (isInitializing) return;

  isInitializing = true;
  latestInitError = null;
  currentClientState = "initializing";
  broadcastStatus();

  try {
    client = createClient();

    client.on("qr", async (qr) => {
      latestQrCode = await QRCode.toDataURL(qr, { margin: 1, width: 320 });
      isReady = false;
      isInitializing = false;
      currentClientState = "qr";
      broadcast("qr", { qr: latestQrCode });
      broadcastStatus();
    });

    client.on("authenticated", () => {
      isAuthenticated = true;
      latestQrCode = null;
      currentClientState = "authenticated";
      broadcastStatus();
    });

    client.on("loading_screen", (percent, message) => {
      loadingPercent = Number(percent || 0);
      currentClientState = message || "loading";
      broadcastStatus();
    });

    client.on("change_state", (state) => {
      currentClientState = state || currentClientState;
      broadcastStatus();
    });

    client.on("ready", () => {
      isReady = true;
      isAuthenticated = true;
      isInitializing = false;
      latestQrCode = null;
      latestInitError = null;
      loadingPercent = null;
      currentClientState = "ready";
      broadcastStatus();
      broadcast("chats", { reason: "ready" });
    });

    client.on("message", broadcastMessage);
    client.on("message_create", broadcastMessage);

    client.on("auth_failure", (message) => {
      isReady = false;
      isAuthenticated = false;
      isInitializing = false;
      latestInitError = message || "Authentication failed";
      currentClientState = "auth_failure";
      broadcastStatus();
    });

    client.on("disconnected", (reason) => {
      isReady = false;
      isAuthenticated = false;
      isInitializing = false;
      latestQrCode = null;
      lastDisconnectReason = reason || "unknown";
      currentClientState = "disconnected";
      broadcastStatus();

      setTimeout(() => {
        initializeClient().catch((error) => {
          latestInitError = error.message;
          broadcastStatus();
        });
      }, 3000).unref?.();
    });

    await client.initialize();
  } catch (error) {
    isReady = false;
    isInitializing = false;
    latestInitError = error.message;
    currentClientState = "error";
    broadcastStatus();
  }
}

app.get("/api/status", (_req, res) => {
  res.json(statusPayload());
});

app.get("/api/qr", (_req, res) => {
  res.json({ qr: latestQrCode, status: statusPayload() });
});

app.get("/api/events", (req, res) => {
  res.writeHead(200, {
    "Content-Type": "text/event-stream",
    "Cache-Control": "no-cache, no-transform",
    Connection: "keep-alive",
    "X-Accel-Buffering": "no",
  });

  res.write(": connected\n\n");
  eventClients.add(res);
  sendEvent(res, "status", statusPayload());
  if (latestQrCode) sendEvent(res, "qr", { qr: latestQrCode });

  const keepAlive = setInterval(() => {
    res.write(": keepalive\n\n");
  }, 25000);

  req.on("close", () => {
    clearInterval(keepAlive);
    eventClients.delete(res);
  });
});

app.get("/api/chats", async (req, res) => {
  if (!ensureReady(res)) return;

  try {
    const limit = normalizeLimit(req.query.limit, 80, 200);
    const chats = await client.getChats();
    const serialized = chats
      .filter(isNormalChat)
      .map(serializeChat)
      .sort((a, b) => b.lastActivityTimestamp - a.lastActivityTimestamp)
      .slice(0, limit);

    res.json({ chats: serialized });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get("/api/chats/:id/messages", async (req, res) => {
  if (!ensureReady(res)) return;

  try {
    const chatId = req.params.id;
    const limit = normalizeLimit(req.query.limit, 80, 500);
    const chat = await client.getChatById(chatId);
    const messages = await chat.fetchMessages({ limit });
    const serialized = messages
      .map(serializeMessage)
      .sort((a, b) => a.timestamp - b.timestamp);

    res.json({ messages: serialized });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.post("/api/chats/:id/send", async (req, res) => {
  if (!ensureReady(res)) return;

  const chatId = req.params.id;
  const body = String(req.body?.body || "").trim();
  if (!body) {
    res.status(400).json({ error: "Message body is required." });
    return;
  }

  try {
    const sent = await client.sendMessage(chatId, body);
    broadcastMessage(sent);
    res.json({ message: serializeMessage(sent) });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.post("/api/logout", async (_req, res) => {
  try {
    if (client) {
      await client.logout();
      await client.destroy();
    }
  } catch (error) {
    latestInitError = error.message;
  }

  client = null;
  isReady = false;
  isAuthenticated = false;
  isInitializing = false;
  latestQrCode = null;
  currentClientState = "logged_out";
  broadcastStatus();
  initializeClient();
  res.json({ ok: true, status: statusPayload() });
});

app.post("/api/session/reset", async (_req, res) => {
  try {
    if (client) {
      await client.destroy();
    }
  } catch (error) {
    latestInitError = error.message;
  }

  client = null;
  isReady = false;
  isAuthenticated = false;
  isInitializing = false;
  latestQrCode = null;
  currentClientState = "resetting";

  const sessionPath = path.join(AUTH_DIR, `session-${CLIENT_ID}`);
  if (fs.existsSync(sessionPath)) {
    fs.rmSync(sessionPath, { recursive: true, force: true });
  }

  broadcastStatus();
  initializeClient();
  res.json({ ok: true, status: statusPayload() });
});

app.get("/healthz", (_req, res) => {
  res.json({ ok: true, status: statusPayload() });
});

app.listen(PORT, () => {
  console.log(`WhatsApp Web bridge listening on ${PORT}`);
  initializeClient();
});

process.on("SIGINT", async () => {
  if (client) await client.destroy();
  process.exit(0);
});

process.on("SIGTERM", async () => {
  if (client) await client.destroy();
  process.exit(0);
});
