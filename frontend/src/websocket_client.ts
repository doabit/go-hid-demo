interface ClientOptions {
  autoConnect: boolean;
  protocols?: string[];
  timeout: number;
  reconnectionAttempts: number;
  reconnectionDelay: number;
}

type EventFunc = (event: any) => void;

const enum WebSocketEventEnum {
  open = "open",
  close = "close",
  error = "error",
  message = "message",
  reconnectAttempt = "reconnectAttempt",
  reconnectFailed = "reconnectFailed",
}

const getDefaultOptions = (): ClientOptions => ({
  autoConnect: true,
  timeout: 20_000,
  reconnectionAttempts: Infinity,
  reconnectionDelay: 5_000,
});

const getEmptyEventsMap = (): Record<WebSocketEventEnum, EventFunc[]> => ({
  [WebSocketEventEnum.open]: [],
  [WebSocketEventEnum.close]: [],
  [WebSocketEventEnum.error]: [],
  [WebSocketEventEnum.message]: [],
  [WebSocketEventEnum.reconnectAttempt]: [],
  [WebSocketEventEnum.reconnectFailed]: [],
});

class WebsocketClient {
  private url: string;
  private options: ClientOptions;
  private websocket: WebSocket | null;
  private events = getEmptyEventsMap();
  private static instance: WebsocketClient | null = null;
  private timer: any = null;
  private reconnectionAttempts = 0;

  public static getInstance(url: string, options: Partial<ClientOptions> = {}) {
    if (!WebsocketClient.instance) {
      WebsocketClient.instance = new WebsocketClient(url, options);
    }
    return WebsocketClient.instance;
  }

  constructor(url: string, options: Partial<ClientOptions>) {
    this.url = url;
    this.options = {
      ...getDefaultOptions(),
      ...options,
    };

    if (this.options.autoConnect) {
      this.connect();
    }
  }

  public connect(resetReconnectionAttempts = true) {
    // 手动调用默认重置重连，但是内部调用不需要清空
    if (resetReconnectionAttempts) {
      this.reconnectionAttempts = 0;
    }
    this.websocket = new WebSocket(this.url, this.options.protocols);
    this.setTimer();

    this.websocket.onopen = (event) => {
      this.clearTimer();
      this.emit(WebSocketEventEnum.open, event);
    };

    this.websocket.onmessage = (event) => {
      this.emit(WebSocketEventEnum.message, event);
    };

    this.websocket.onerror = (event) => {
      this.emit(WebSocketEventEnum.error, event);
      this.reconnect(event);
    };

    this.websocket.onclose = (event) => {
      this.emit(WebSocketEventEnum.close, event);
      this.reconnect(event);
    };
  }

  public on(name: WebSocketEventEnum, listener: EventFunc) {
    this.events[name].push(listener);
  }

  public off(name?: WebSocketEventEnum, listener?: EventFunc) {
    if (!name) {
      this.events = getEmptyEventsMap();
      return;
    }

    if (!listener) {
      this.events[name] = [];
      return;
    }

    const index = this.events[name].findIndex((fn) => fn === listener);
    if (index > -1) {
      this.events[name].splice(index, 1);
    }
  }

  public send(
    data: string | ArrayBuffer | SharedArrayBuffer | Blob | ArrayBufferView
  ) {
    if (this.websocket?.readyState === WebSocket.OPEN) {
      this.websocket.send(data);
    } else {
      console.error('WebSocket is not in OPEN state');
    }
  }

  public disconnect() {
    this.reconnectionAttempts = -1;
    this.websocket?.close(1_000, "Normal Closure");
  }

  private reconnect(event: Event) {
    // -1时不需要重连
    if (this.reconnectionAttempts === -1) {
      this.websocket = null;
      return;
    }

    // 疑问2
    if (this.websocket?.readyState !== WebSocket.CLOSED) {
      return;
    }

    this.websocket = null;
    this.reconnectionAttempts++;
    this.emit(WebSocketEventEnum.reconnectAttempt, this.reconnectionAttempts);
    if (
      !Number.isFinite(this.options.reconnectionAttempts) ||
      this.reconnectionAttempts <= this.options.reconnectionAttempts
    ) {
      setTimeout(() => {
        this.connect(false);
      }, this.options.reconnectionDelay);
      return;
    }

    this.emit(WebSocketEventEnum.reconnectFailed, event);
  }

  private emit(name: WebSocketEventEnum, event: any) {
    this.events[name].forEach((listener) => listener(event));
  }

  private setTimer() {
    this.clearTimer();
    this.timer = setTimeout(() => {
      // 疑问1
      this.websocket?.close();
    }, this.options.timeout);
  }

  private clearTimer() {
    this.timer !== null && clearTimeout(this.timer);
  }
}
