class ReconnectingWebSocket {
    constructor(url, protocols, options = {}) {
      this.url = url;
      this.protocols = protocols;
      this.options = options;

      this.ws = null;
      this.reconnectTimeout = null;
      this.reconnectDelay = options.reconnectDelay || 1000;
      this.maxReconnectAttempts = options.maxReconnectAttempts || Infinity;
      this.attempts = 0;

      this.connect();
    }

    connect() {
      this.ws = new WebSocket(this.url, this.protocols);
      this.ws.addEventListener('open', (event) => this.onOpen(event));
      this.ws.addEventListener('message', (event) => this.onMessage(event));
      this.ws.addEventListener('close', (event) => this.onClose(event));
      this.ws.addEventListener('error', (event) => this.onError(event));
    }

    onOpen(event) {
      this.attempts = 0;
      if (this.options.onOpen) {
        this.options.onOpen(event);
      }
    }

    onMessage(event) {
      if (this.options.onMessage) {
        this.options.onMessage(event);
      }
    }

    onClose(event) {
      if (this.attempts < this.maxReconnectAttempts) {
        this.attempts++;
        this.reconnectTimeout = setTimeout(() => this.connect(), this.reconnectDelay);
      }

      if (this.options.onClose) {
        this.options.onClose(event);
      }
    }

    onError(event) {
      if (this.options.onError) {
        this.options.onError(event);
      }
    }

    send(data) {
      if (this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(data);
      } else {
        console.error('WebSocket is not in OPEN state');
      }
    }

    close(code, reason) {
      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout);
        this.reconnectTimeout = null;
      }

      this.ws.close(code, reason);
    }
  }
