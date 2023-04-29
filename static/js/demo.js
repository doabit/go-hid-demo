const ws = new ReconnectingWebSocket("ws://localhost:8080/ws", null, {
  reconnectDelay: 2000,
  maxReconnectAttempts: 30,
  onOpen: (event) => {
    console.log('WebSocket connection opened:', event);
  },
  onMessage: (event) => {
    console.log('WebSocket message received:', event);
  },
  onClose: (event) => {
    console.log('WebSocket connection closed:', event);
  },
  onError: (event) => {
    console.log('WebSocket error:', event);
  },
});

// ws.send('Hello, WebSocket!');
// ws.close(1000, 'Normal closure');
