let ws = null;
let playerName = '';

function updateConnectionStatus(connected) {
    const statusEl = document.getElementById('connection-status');
    statusEl.textContent = connected ? 'Connected' : 'Disconnected';
    statusEl.className = connected ? 'connected' : 'disconnected';
}

function addMessage(text, type = 'server') {
    const messageList = document.getElementById('message-list');
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${type}`;
    messageDiv.textContent = text;
    messageList.appendChild(messageDiv);
    messageList.scrollTop = messageList.scrollHeight;
}

function updateBoard(boardText) {
    const boardDisplay = document.getElementById('board-display');
    boardDisplay.textContent = boardText;
}

function updateGameInfo(message) {
    const gameInfo = document.getElementById('game-info');

    if (message.includes('You are player') ||
        message.includes('Current turn:') ||
        message.includes('Game ID:') ||
        message.includes('Match found!')) {
        gameInfo.textContent = message.split('\n')[0];
    }
}

function connect() {
    const nameInput = document.getElementById('player-name');
    playerName = nameInput.value.trim() || 'Guest';

    const serverURL = `ws://localhost:8080/ws?name=${encodeURIComponent(playerName)}`;

    try {
        ws = new WebSocket(serverURL);

        ws.onopen = () => {
            console.log('Connected to server');
            updateConnectionStatus(true);
            document.getElementById('connection-panel').style.display = 'none';
            document.getElementById('game-panel').style.display = 'block';
            addMessage(`Connected as ${playerName}`, 'important');

            document.getElementById('move-input').focus();
        };

        ws.onmessage = (event) => {
            const message = event.data;
            console.log('Received:', message);

            if (message.includes('Current turn:') && message.includes('MOVE LEGEND')) {
                updateBoard(message);
            }

            updateGameInfo(message);

            const messageType = message.toLowerCase().includes('error') ||
                message.toLowerCase().includes('invalid') ? 'error' :
                message.includes('Match found!') ||
                    message.includes('wins') ||
                    message.includes('Draw!') ? 'important' : 'server';

            addMessage(message, messageType);
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            addMessage('Connection error occurred', 'error');
        };

        ws.onclose = () => {
            console.log('Disconnected from server');
            updateConnectionStatus(false);
            addMessage('Disconnected from server', 'error');
            ws = null;
        };

    } catch (error) {
        console.error('Failed to connect:', error);
        addMessage('Failed to connect to server', 'error');
    }
}

function disconnect() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close();
    }
    document.getElementById('connection-panel').style.display = 'block';
    document.getElementById('game-panel').style.display = 'none';
    document.getElementById('message-list').innerHTML = '';
    document.getElementById('board-display').textContent = '';
    document.getElementById('game-info').textContent = '';
    updateConnectionStatus(false);
}

function sendMessage(message) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(message);
        console.log('Sent:', message);
    } else {
        addMessage('Not connected to server', 'error');
    }
}

function sendMove() {
    const moveInput = document.getElementById('move-input');
    const move = moveInput.value.trim().toUpperCase();

    if (move) {
        sendMessage(move);
        moveInput.value = '';
    }
}

function showBoard() {
    sendMessage('board');
}

function showStatus() {
    sendMessage('status');
}

function showHelp() {
    sendMessage('help');
}

document.addEventListener('DOMContentLoaded', () => {
    const moveInput = document.getElementById('move-input');
    const nameInput = document.getElementById('player-name');

    moveInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendMove();
        }
    });

    nameInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            connect();
        }
    });
});
