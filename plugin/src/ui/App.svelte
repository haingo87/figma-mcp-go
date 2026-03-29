<script lang="ts">
  import { onMount } from "svelte";

  let connected = false;
  let fileName = "—";
  let selectionCount = 0;

  const WS_URL = "ws://localhost:1994/ws";
  const RECONNECT_DELAY_MS = 1500;

  let socket: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  function connect() {
    if (socket) socket.close();
    socket = new WebSocket(WS_URL);

    socket.onopen = () => {
      connected = true;
      parent.postMessage({ pluginMessage: { type: "ui-ready" } }, "*");
    };

    socket.onclose = () => {
      connected = false;
      socket = null;
      if (reconnectTimer === null) {
        reconnectTimer = setTimeout(() => {
          reconnectTimer = null;
          connect();
        }, RECONNECT_DELAY_MS);
      }
    };

    socket.onerror = () => {
      connected = false;
    };

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        parent.postMessage({ pluginMessage: { type: "server-request", payload } }, "*");
      } catch {
        // ignore malformed frames
      }
    };
  }

  function handleMessage(event: MessageEvent) {
    const msg = event.data?.pluginMessage;
    if (!msg) return;

    if (msg.type === "plugin-status") {
      fileName = msg.payload.fileName;
      selectionCount = msg.payload.selectionCount;
      return;
    }

    if ("requestId" in msg && socket?.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(msg));
    }
  }

  onMount(() => {
    window.addEventListener("message", handleMessage);
    connect();

    return () => {
      window.removeEventListener("message", handleMessage);
      if (reconnectTimer !== null) clearTimeout(reconnectTimer);
      if (socket) socket.close();
    };
  });
</script>

<div class="container">
  <div class="info-section">
    <div class="info-row">
      <span class="info-label">File</span>
      <span class="info-value">{fileName}</span>
    </div>
    <div class="info-row">
      <span class="info-label">Selection</span>
      <span class="info-value">{selectionCount} node(s)</span>
    </div>
  </div>
  <div class="footer">
    <a
      class="author"
      href="https://github.com/vkhanhqui/figma-mcp-go"
      target="_blank"
    >
      <img
        src="https://avatars.githubusercontent.com/u/64468109?v=4"
        alt="avatar"
      />
      vkhanhqui
    </a>
    <div class="badge" class:connected class:disconnected={!connected}>
      <span class="dot" class:connected></span>
      <span>{connected ? "Connected" : "Disconnected"}</span>
    </div>
  </div>
</div>

<style>
  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    font-size: 12px;
    background: #1e1e1e;
    color: #e0e0e0;
    height: 100vh;
  }

  .container {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 16px;
    gap: 12px;
  }

  .info-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 1;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .info-label {
    color: #888;
  }

  .info-value {
    color: #e0e0e0;
    font-weight: 500;
  }

  .footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .author {
    display: flex;
    align-items: center;
    gap: 6px;
    text-decoration: none;
    color: #888;
    font-size: 11px;
  }

  .author:hover {
    color: #e0e0e0;
  }

  .author img {
    width: 20px;
    height: 20px;
    border-radius: 50%;
  }

  .badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 11px;
    font-weight: 600;
  }

  .badge.connected {
    background: #1a472a;
    color: #4ade80;
  }

  .badge.disconnected {
    background: #3a1a1a;
    color: #f87171;
  }

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #f87171;
  }

  .dot.connected {
    background: #4ade80;
  }
</style>
