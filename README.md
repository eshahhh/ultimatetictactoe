# ultimate-tictactoe

I have created the server that both the cli client and a web client can connect to (most focus is on the webclient though)
There is also a proper matchmaking queuing system and the web client properly connects to the server to be able to join matches, so a web client player can match against a cmd player. 

For some reason nest is not working which means I can't deploy this project properly thats why I had to use a video demo :( 

Feel free to download and check everything yourself!

Server
```
go run cmd/server/main.go
```

Frontend
```
cd web-client
python -m http.server 8080
```

Navigate to localhost:8080 and open join the game from different tabs or wait to get matched to a player.
