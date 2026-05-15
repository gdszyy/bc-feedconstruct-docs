import type {
  ControlFrame,
  SubscribeScope,
  TypedEnvelope,
} from "@/contract/events";

export type ConnectionState =
  | "Disconnected"
  | "Connecting"
  | "Open"
  | "Degraded"
  | "Reconnecting"
  | "Closed";

export const ORIGIN_REJECTED_CLOSE_CODE = 4401;

export interface WebSocketLike {
  send(data: string): void;
  close(code?: number, reason?: string): void;
  onopen: ((ev?: unknown) => void) | null;
  onmessage: ((ev: { data: string }) => void) | null;
  onclose: ((ev: { code: number; reason?: string }) => void) | null;
  onerror: ((ev?: unknown) => void) | null;
  readyState: number;
}

export type WebSocketFactory = (url: string) => WebSocketLike;

export interface TransportError {
  code: "ORIGIN_REJECTED" | "MAX_RECONNECT" | "UNKNOWN";
  message: string;
  retriable: boolean;
}

export interface TransportOptions {
  url: string;
  reconnectMinDelayMs?: number;
  reconnectMaxDelayMs?: number;
  webSocketFactory?: WebSocketFactory;
  scheduleTimeout?: (cb: () => void, ms: number) => unknown;
  cancelTimeout?: (handle: unknown) => void;
}

const DEFAULT_MIN_DELAY_MS = 1000;
const DEFAULT_MAX_DELAY_MS = 30000;

export class Transport {
  private readonly url: string;
  private readonly reconnectMinDelay: number;
  private readonly reconnectMaxDelay: number;
  private readonly wsFactory: WebSocketFactory;
  private readonly scheduleTimeout: (cb: () => void, ms: number) => unknown;
  private readonly cancelTimeout: (handle: unknown) => void;

  private state: ConnectionState = "Disconnected";
  private ws: WebSocketLike | null = null;
  private subscriptions = new Map<string, SubscribeScope>();
  private reconnectAttempt = 0;
  private receivedSinceOpen = false;
  private pendingReconnect: unknown = null;
  private stopped = false;

  private stateListeners = new Set<(s: ConnectionState) => void>();
  private messageListeners = new Set<(env: TypedEnvelope) => void>();
  private errorListeners = new Set<(e: TransportError) => void>();

  constructor(opts: TransportOptions) {
    this.url = opts.url;
    this.reconnectMinDelay = opts.reconnectMinDelayMs ?? DEFAULT_MIN_DELAY_MS;
    this.reconnectMaxDelay = opts.reconnectMaxDelayMs ?? DEFAULT_MAX_DELAY_MS;
    this.wsFactory = opts.webSocketFactory ?? defaultWebSocketFactory;
    this.scheduleTimeout =
      opts.scheduleTimeout ?? ((cb, ms) => setTimeout(cb, ms));
    this.cancelTimeout =
      opts.cancelTimeout ??
      ((h) => clearTimeout(h as ReturnType<typeof setTimeout>));
  }

  getState(): ConnectionState {
    return this.state;
  }

  connect(): void {
    if (this.stopped) return;
    if (this.state === "Connecting" || this.state === "Open") return;

    this.transition("Connecting");
    const ws = this.wsFactory(this.url);
    this.ws = ws;

    this.receivedSinceOpen = false;
    ws.onopen = () => {
      this.transition("Open");
      this.replaySubscriptions();
    };
    ws.onmessage = (ev) => this.handleMessage(ev.data);
    ws.onclose = (ev) => this.handleClose(ev.code);
    ws.onerror = () => {
      /* close event always follows; error itself is non-fatal */
    };
  }

  close(): void {
    this.stopped = true;
    if (this.pendingReconnect !== null) {
      this.cancelTimeout(this.pendingReconnect);
      this.pendingReconnect = null;
    }
    if (this.ws) this.ws.close();
    this.transition("Closed");
  }

  subscribe(scope: SubscribeScope): void {
    const key = serializeScope(scope);
    this.subscriptions.set(key, scope);
    this.sendFrame({ op: "subscribe", scope });
  }

  unsubscribe(scope: SubscribeScope): void {
    const key = serializeScope(scope);
    this.subscriptions.delete(key);
    this.sendFrame({ op: "unsubscribe", scope });
  }

  onState(handler: (s: ConnectionState) => void): () => void {
    this.stateListeners.add(handler);
    return () => this.stateListeners.delete(handler);
  }

  onMessage(handler: (env: TypedEnvelope) => void): () => void {
    this.messageListeners.add(handler);
    return () => this.messageListeners.delete(handler);
  }

  onError(handler: (e: TransportError) => void): () => void {
    this.errorListeners.add(handler);
    return () => this.errorListeners.delete(handler);
  }

  private handleMessage(raw: string): void {
    let parsed: TypedEnvelope;
    try {
      parsed = JSON.parse(raw) as TypedEnvelope;
    } catch {
      // M02/M16 will own parse-error routing once wired up.
      return;
    }
    if (!this.receivedSinceOpen) {
      this.receivedSinceOpen = true;
      this.reconnectAttempt = 0;
    }
    this.messageListeners.forEach((l) => l(parsed));
  }

  private handleClose(code: number): void {
    this.ws = null;

    if (code === ORIGIN_REJECTED_CLOSE_CODE) {
      this.stopped = true;
      this.transition("Closed");
      this.emitError({
        code: "ORIGIN_REJECTED",
        message: "Origin rejected by BFF (4401)",
        retriable: false,
      });
      return;
    }

    if (this.stopped) {
      this.transition("Closed");
      return;
    }

    this.scheduleReconnect();
  }

  private scheduleReconnect(): void {
    this.transition("Reconnecting");
    const attempt = this.reconnectAttempt++;
    const delay = Math.min(
      this.reconnectMaxDelay,
      this.reconnectMinDelay * 2 ** attempt,
    );
    this.pendingReconnect = this.scheduleTimeout(() => {
      this.pendingReconnect = null;
      this.connect();
    }, delay);
  }

  private replaySubscriptions(): void {
    for (const scope of this.subscriptions.values()) {
      this.sendFrame({ op: "subscribe", scope });
    }
  }

  private sendFrame(frame: ControlFrame): void {
    if (!this.ws || this.state !== "Open") return;
    this.ws.send(JSON.stringify(frame));
  }

  private transition(next: ConnectionState): void {
    if (this.state === next) return;
    this.state = next;
    this.stateListeners.forEach((l) => l(next));
  }

  private emitError(e: TransportError): void {
    this.errorListeners.forEach((l) => l(e));
  }
}

function serializeScope(scope: SubscribeScope): string {
  const norm: Record<string, string[]> = {};
  if (scope.match_ids) norm.match_ids = [...scope.match_ids].sort();
  if (scope.sport_ids) norm.sport_ids = [...scope.sport_ids].sort();
  if (scope.tournament_ids)
    norm.tournament_ids = [...scope.tournament_ids].sort();
  return JSON.stringify(norm);
}

function defaultWebSocketFactory(url: string): WebSocketLike {
  return new WebSocket(url) as unknown as WebSocketLike;
}
