# ultimate-tictactoe

I have created the server that both the cli client and a web client can connect to (most focus is on the webclient though)
There is also a proper matchmaking queuing system and the web client properly connects to the server to be able to join matches, so a web client player can match against a cmd player. 

To test it out, navigate to the [https://ultimatetictactoe-eshahhh.vercel.app/] and then either open the link again in a new tab or send it to your friend. The server should automatically matchmake the two together (unless someone else joins at the exact same time lol) and you should be able to play a match of UTTT!

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
