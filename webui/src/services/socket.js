import {getToken} from "../store/auth";

export function initializeWebSocket(app) {
    const socket = new WebSocket(`ws://localhost:3000/ws?token=${getToken()}`);

    socket.onopen = () => {
        console.log('WebSocket connection established');
    };

    socket.onmessage = (event) => {
        const message = JSON.parse(event.data);
        console.log('New message:', message);

        app.messages.push({
            message: message.message,
            isSent: false,
        });
    };

    socket.onclose = () => {
        console.log('WebSocket connection closed');
    };

    return socket;
}