# UGN - Ultimate Game Notation

UGN (Ultimate Game Notation) is a compact notation system for recording Ultimate Tic-Tac-Toe games, inspired by PGN (Portable Game Notation) used in chess.

## Basic Move Notation

A basic move consists of:
- **Board identifier**: A letter from A-I representing the 3x3 grid of boards
- **Cell number**: A digit from 1-9 representing the position within that board

Example: `A5` means "play in cell 5 of board A"

## Board Layout

The boards are arranged in a 3x3 grid:
```
A | B | C
---------
D | E | F
---------
G | H | I
```

Each board contains cells numbered 1-9:
```
1 | 2 | 3
---------
4 | 5 | 6
---------
7 | 8 | 9
```

## Game Flow Notation

Moves are written in sequence, separated by spaces. Each move sends the opponent to a specific board (determined by the cell number played).

Example: `A5 E8`
- X plays cell 5 in board A
- This sends O to board E
- O plays cell 8 in board E

## Special Symbols

UGN uses special symbols to indicate important game events:

- `!` - **Small board win**: The move wins a small board
- `/` - **Small board draw**: The move causes a draw on a small board  
- `%` - **Game draw**: The move results in a game draw
- `#` - **Game win**: The move wins the entire game

### Examples with Special Symbols

`A5/ E9!` 
- X plays A5, causing board A to draw
- O plays E9, winning board E

`I4/ D2#`
- X plays I4, causing board I to draw  
- O plays D2, winning the entire game

## Complete Game Example

```
A5 E8 
H3! C6/ 
F1 A7 
B4! G2 
I5/ D8#
```

This represents:
1. X plays A5, O plays E8
2. X plays H3 and wins board H, O plays C6 and draws board C
3. X plays F1, O plays A7
4. X plays B4 and wins board B, O plays G2
5. X plays I5 and draws board I, O plays D8 and wins the game

## File Format

UGN files follow this format:

```
[GameID "default"]
[Date "2025-06-29"]
[Time "15:39:29"]
[PlayerX "Player1"]
[PlayerO "Player2"]
[Result "X"]
[Comment "X wins by resignation"]

A1 A3
C4 D5
E5 E2
B1
1-0

```

### Metadata Fields
- **GameID**: Unique identifier for the game.
- **Date**: Date when the game was played (YYYY-MM-DD format).
- **Time**: Time when the game started (HH:MM:SS format).
- **PlayerX**: Name or address of the player playing as X.
- **PlayerO**: Name or address of the player playing as O.
- **Result**: Final result - "X", "O", or "Draw".
- **Comment**: Optional comment describing the game result (e.g., "X wins by resignation").

## File Naming Convention

UGN files are stored with the following naming pattern:
`YYYYMMDD_HHMMSS_gameID.ugn`

Example: `20250625_143022_abc123def.ugn`
