import {getToken} from "../store/auth";

export function initializeWebSocket(app, groupId) {
    let socketUrl = `ws://localhost:3000/ws?token=${getToken()}`;

    if (groupId !== 0) {
        socketUrl += `&groupId=${groupId}`;
    }

    const socket = new WebSocket(socketUrl);

    socket.onopen = () => {
        console.log('WebSocket connection established');
    };

    socket.onmessage = (event) => {
        const message = JSON.parse(event.data);
        console.log('New message:', message.message);

        app.messages.push({
            id:message.id,
            message: message.message,
            isPhoto:message.isPhoto,
            username: message.username ?? app.userInfo.Username,
            createdAt: message.createdAt,
            isReceived: message.isReceived,
            isSent: false,
            isRead: message.isRead,
        });
    };

    socket.onclose = () => {
        console.log('WebSocket connection closed');
    };

    return socket;
}