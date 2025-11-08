let ws = null;
let playerName = '';
let gameState = null;

const BOARD_LETTERS = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I'];

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

function updateGameInfo(state) {
    if (!state) return;

    document.getElementById('game-id').textContent = `Game ID: ${state.game_id}`;
    document.getElementById('your-symbol').textContent = `You: ${state.your_symbol}`;
    document.getElementById('player-x').textContent = `X: ${state.player_x_name}`;
    document.getElementById('player-o').textContent = `O: ${state.player_o_name}`;
    document.getElementById('current-turn').textContent = `Current Turn: ${state.current_turn}`;

    let statusText = state.game_status === 'finished' ?
        `Game Over - ${state.winner}` :
        state.is_your_turn ? 'Your Turn!' : 'Opponent\'s Turn';
    document.getElementById('game-status').textContent = `Status: ${statusText}`;
}

function createBoard() {
    const ultimateBoard = document.getElementById('ultimate-board');
    ultimateBoard.innerHTML = '';

    for (let i = 0; i < 9; i++) {
        const smallBoard = document.createElement('div');
        smallBoard.className = 'small-board';
        smallBoard.dataset.boardIndex = i;

        for (let j = 0; j < 9; j++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.dataset.boardIndex = i;
            cell.dataset.cellIndex = j;
            cell.addEventListener('click', () => handleCellClick(i, j));
            smallBoard.appendChild(cell);
        }

        ultimateBoard.appendChild(smallBoard);
    }
}

function renderBoard(state) {
    if (!state || !state.board) return;

    gameState = state;

    for (let boardIndex = 0; boardIndex < 9; boardIndex++) {
        const smallBoard = document.querySelector(`.small-board[data-board-index="${boardIndex}"]`);
        const boardData = state.board.boards[boardIndex];
        const boardState = state.board.board_states[boardIndex];

        smallBoard.className = 'small-board';

        if (boardState === 'X') {
            smallBoard.classList.add('won-x');
        } else if (boardState === 'O') {
            smallBoard.classList.add('won-o');
        } else if (boardState === 'draw') {
            smallBoard.classList.add('draw');
        }

        if (state.active_board === boardIndex && state.game_status === 'in_progress') {
            smallBoard.classList.add('active');
        } else if (state.active_board !== -1 && state.active_board !== boardIndex && state.game_status === 'in_progress') {
            smallBoard.classList.add('disabled');
        }

        let existingWinner = smallBoard.querySelector('.board-winner');
        if (boardState !== 'undecided' && !existingWinner) {
            const winnerDiv = document.createElement('div');
            winnerDiv.className = `board-winner ${boardState.toLowerCase()}`;
            winnerDiv.textContent = boardState === 'draw' ? 'D' : boardState;
            smallBoard.appendChild(winnerDiv);
        }

        for (let cellIndex = 0; cellIndex < 9; cellIndex++) {
            const cell = smallBoard.querySelector(`.cell[data-cell-index="${cellIndex}"]`);
            const cellValue = boardData.cells[cellIndex];

            cell.className = 'cell';
            cell.textContent = cellValue || '';

            if (cellValue) {
                cell.classList.add('filled', cellValue.toLowerCase());
            }

            const canPlay = state.is_your_turn &&
                state.game_status === 'in_progress' &&
                (state.active_board === -1 || state.active_board === boardIndex) &&
                boardState === 'undecided' &&
                !cellValue;

            if (!canPlay) {
                cell.classList.add('disabled');
            }
        }
    }
}

function handleCellClick(boardIndex, cellIndex) {
    if (!gameState || !gameState.is_your_turn || gameState.game_status !== 'in_progress') {
        addMessage('It\'s not your turn!', 'error');
        return;
    }

    if (gameState.active_board !== -1 && gameState.active_board !== boardIndex) {
        addMessage(`You must play on board ${BOARD_LETTERS[gameState.active_board]}!`, 'error');
        return;
    }

    if (gameState.board.board_states[boardIndex] !== 'undecided') {
        addMessage('This board is already finished!', 'error');
        return;
    }

    if (gameState.board.boards[boardIndex].cells[cellIndex]) {
        addMessage('This cell is already taken!', 'error');
        return;
    }

    const moveStr = BOARD_LETTERS[boardIndex] + (cellIndex + 1);
    sendMove(moveStr);
}

function updateUGNNotation(moves) {
    const ugnMovesDiv = document.getElementById('ugn-moves');
    ugnMovesDiv.innerHTML = '';

    if (!moves || moves.length === 0) {
        ugnMovesDiv.textContent = 'No moves yet';
        return;
    }

    for (let i = 0; i < moves.length; i += 2) {
        const pairDiv = document.createElement('div');
        pairDiv.className = 'ugn-move-pair';

        const moveNum = document.createElement('span');
        moveNum.className = 'ugn-move-number';
        moveNum.textContent = `${Math.floor(i / 2) + 1}.`;
        pairDiv.appendChild(moveNum);

        const xMove = document.createElement('span');
        xMove.className = 'ugn-move';
        if (i === moves.length - 1) {
            xMove.classList.add('current');
        }
        xMove.textContent = moves[i];
        pairDiv.appendChild(xMove);

        if (i + 1 < moves.length) {
            const oMove = document.createElement('span');
            oMove.className = 'ugn-move';
            if (i + 1 === moves.length - 1) {
                oMove.classList.add('current');
            }
            oMove.textContent = moves[i + 1];
            pairDiv.appendChild(oMove);
        }

        ugnMovesDiv.appendChild(pairDiv);
    }

    const container = document.getElementById('ugn-moves-container');
    container.scrollTop = container.scrollHeight;
}

function handleMessage(data) {
    try {
        const msg = JSON.parse(data);

        switch (msg.type) {
            case 'welcome':
                addMessage(msg.payload.message, 'important');
                break;

            case 'game_state':
                updateGameInfo(msg.payload);
                renderBoard(msg.payload);
                updateUGNNotation(msg.payload.ugn_moves);
                break;

            case 'move':
                addMessage(`${msg.payload.player_name} (${msg.payload.player_symbol}) played ${msg.payload.move}`, 'info');
                break;

            case 'error':
                addMessage(`Error: ${msg.payload.message}`, 'error');
                break;

            case 'info':
                addMessage(msg.payload.message, 'info');
                break;

            case 'game_over':
                addMessage(msg.payload.message, 'important');
                break;

            case 'draw_offer':
                addMessage(msg.payload.message, 'important');
                showDrawOfferButtons();
                break;

            default:
                console.log('Unknown message type:', msg.type);
        }
    } catch (e) {
        console.error('Failed to parse message:', e);
        addMessage(data, 'server');
    }
}

function showDrawOfferButtons() {
    const controls = document.getElementById('controls');

    const existingButtons = controls.querySelectorAll('.draw-response');
    existingButtons.forEach(btn => btn.remove());

    const acceptBtn = document.createElement('button');
    acceptBtn.textContent = 'Accept Draw';
    acceptBtn.className = 'draw-response';
    acceptBtn.onclick = acceptDraw;
    controls.insertBefore(acceptBtn, controls.firstChild);

    const declineBtn = document.createElement('button');
    declineBtn.textContent = 'Decline Draw';
    declineBtn.className = 'draw-response';
    declineBtn.onclick = declineDraw;
    controls.insertBefore(declineBtn, controls.firstChild);
}

function removeDrawOfferButtons() {
    const buttons = document.querySelectorAll('.draw-response');
    buttons.forEach(btn => btn.remove());
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

            createBoard();

            document.getElementById('move-input').focus();
        };

        ws.onmessage = (event) => {
            handleMessage(event.data);
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
        ws.send('quit');
        ws.close();
    }
    document.getElementById('connection-panel').style.display = 'block';
    document.getElementById('game-panel').style.display = 'none';
    document.getElementById('message-list').innerHTML = '';
    gameState = null;
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

function sendMove(moveStr) {
    if (!moveStr) {
        const moveInput = document.getElementById('move-input');
        moveStr = moveInput.value.trim().toUpperCase();
        moveInput.value = '';
    }

    if (moveStr) {
        sendMessage(moveStr);
    }
}

function showStatus() {
    sendMessage('status');
}

function resign() {
    if (confirm('Are you sure you want to resign?')) {
        sendMessage('R');
    }
}

function offerDraw() {
    if (confirm('Offer a draw to your opponent?')) {
        sendMessage('DRAW');
    }
}

function acceptDraw() {
    sendMessage('ACCEPT_DRAW');
    removeDrawOfferButtons();
}

function declineDraw() {
    sendMessage('DECLINE_DRAW');
    removeDrawOfferButtons();
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
