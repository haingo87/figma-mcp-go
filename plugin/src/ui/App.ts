// App.ts — Vanilla TypeScript plugin UI
// Holds the WebSocket connection to the Go server and bridges
// messages between the WS and plugin core (code.ts) via postMessage.
// NO React. NO framework. Compiles to a single inlined script.

const WS_URL = "ws://localhost:1994/ws";
const RECONNECT_DELAY_MS = 1500;

interface PluginStatus {
  fileName: string;
  selectionCount: number;
}

let socket: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let connected = false;

// ── DOM refs ──────────────────────────────────────────────────────────────────

function el<T extends HTMLElement>(id: string): T {
  return document.getElementById(id) as T;
}

// ── UI update helpers ─────────────────────────────────────────────────────────

function setConnected(state: boolean): void {
  connected = state;
  const dot = el("dot");
  const badge = el("badge-text");
  const badge_wrap = el("status-badge");
  dot.className = state ? "dot connected" : "dot";
  badge.textContent = state ? "Connected" : "Disconnected";
  badge_wrap.className = state ? "badge connected" : "badge disconnected";
}

function setStatus(status: PluginStatus): void {
  el("file-name").textContent = status.fileName;
  el("selection-count").textContent = `${status.selectionCount} node(s)`;
}

// ── WebSocket bridge ──────────────────────────────────────────────────────────

function connect(): void {
  if (socket) {
    socket.close();
  }

  socket = new WebSocket(WS_URL);

  socket.onopen = () => {
    setConnected(true);
    // Tell plugin core the UI is ready — it will respond with current status
    parent.postMessage({ pluginMessage: { type: "ui-ready" } }, "*");
  };

  socket.onclose = () => {
    setConnected(false);
    socket = null;
    if (reconnectTimer === null) {
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, RECONNECT_DELAY_MS);
    }
  };

  socket.onerror = () => {
    setConnected(false);
  };

  // Messages from Go server → forward to plugin core via postMessage
  socket.onmessage = (event: MessageEvent) => {
    try {
      const payload = JSON.parse(event.data);
      parent.postMessage(
        { pluginMessage: { type: "server-request", payload } },
        "*"
      );
    } catch {
      // ignore malformed frames
    }
  };
}

// Messages from plugin core → forward to Go server via WebSocket
window.addEventListener("message", (event: MessageEvent) => {
  const msg = event.data?.pluginMessage;
  if (!msg) return;

  if (msg.type === "plugin-status") {
    setStatus(msg.payload as PluginStatus);
    return;
  }

  // Responses from code.ts carry a requestId — forward them to the server
  if ("requestId" in msg) {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(msg));
    }
  }
});

// ── Boot ──────────────────────────────────────────────────────────────────────

connect();
