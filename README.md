# GoApp

An application for Go knowledge assessment.

### Run the application

Build everything:
```
make all
```
Run the server:
```
./bin/server
```
Run the client:
```
./bin/client -n 50 -v
```

NOTES:
- `-v` client flag starts verbose logging with output the file `client.log`
- To run the client e.g. with 50 connections and verbose logs run: `./bin/client -n 50 -v`
- To enable csrf protection to the server run: `./bin/server -csrf`

## Description

This is a web application that utilises websockets. A client connects on `localhost:8080` and has three options: a) `open` a websocket connection that reads values from a counter b) `close` the websocket and c) reset the counter to zero. The counter is feeded from a random string generator. On WS session termination, statistics for terminated session are printed.

The application is compiled by running `make` in the root folder and the final binaries are found in the `bin/` folder.

The application has some problems described below that need to be addressed plus some new features that need implementation.

## Problems

### #1

The server prints statistics for each WS sessioned closed but it seems to only count one message while there are more send to each WS session, e.g.

```
2024/03/29 18:28:38 stats.go:11: session a938e316-8536-46e6-8633-bd309fbcf579 has received 1 messages
```

### #2

A more then normal memory usage is observed after many WS sessions which needs investigation.

### #3

A cross-site request forgery is reported by a security audit which needs fixing.

## New features

### A

Modify the random string generator to generate only hex values and verify its accuracy and resource usage by creating a test and a benchmark run.

### B

Extent the API to also return the Hex value in WS connection. I.e. a browser that open a connection to `localhost:8080` should see the HEX values.

E.g.

```
OPEN
RESPONSE: {"iteration":1,"value":"822876EF10"}
RESPONSE: {"iteration":2,"value":"215100491D"}
RESPONSE: {"iteration":3,"value":"05DCC3B6AB"}
CLOSE
```

### C

Create a command line client as a separate application that opens a requested number of sessions simultaneously.

This should be a separate executable generated with the `make` command in the `bin/` folder that will accept an argument with the number of parallel connections that will open on the server. The server part must be modified in a way to support multiple parallel connections and should still print valid statistics for each connection.

E.g.

```
$ ./bin/client -n 3
[conn #0] iteration: 1, value: 66D53ED788
[conn #1] iteration: 1, value: 66D53ED788
[conn #2] iteration: 1, value: 66D53ED788
...
```
