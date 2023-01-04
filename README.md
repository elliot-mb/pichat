# PiChat
A tiny chat app self-hosted via my raspberry pi; written in go, based off an exercise given to me for my concurrent programming classes. 
## Start guide
If you want to try it out for yourself.
### Prerequisites
- [Golang 1.16 or later.](https://go.dev/doc/install)
### Start the client
- Clone this repository
- Enter the top level directory of this project
- Run the client (``go run client/client.go``)\
  Here you can use optional flags like ``-uname=<username>`` to enter an alias (in place of ``<username>``)\
  By default the flag ``-server=<address:port>`` is pointed at my raspberry pi's DDNS, but feel free to host this server locally or elsewhere; in which case you will need to connect by setting this flag to where you decide to host it. 
  
If you see an error about connection refusal when connecting without changing the ``-server`` flag, **the server may be down**. 

### Start the server
- Clone this repository
- Enter the top level directory
- Run the server (``go run server/server.go -port=<port>``)\
  The port flag provides information to the server as to which port to listen on
  
Server refuses all HTTP/S connections as a 'security' measure; it only allows connections from the ``client.go`` program. Or any program that can spoof the behaviour of ``client.go`` by reproducing it's control codes. 
