package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WS_ENDPOINT = "ws://localhost:8080/goapp/ws"
	LOG_FILE    = "client.log"
)

var (
	verbose = false
	logger  *slog.Logger
)

func connect(c int, s chan<- int, exitChan <-chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			logger.Debug(fmt.Sprintf("panic recovery: %v", r))
			s <- c
		}
	}()
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, _, err1 := dialer.Dial(WS_ENDPOINT, http.Header{})
	if err1 != nil {
		logger.Debug(fmt.Sprintf("client %d could not connect", c))
		s <- c
		return
	}
	logger.Debug(fmt.Sprintf("client %d connected", c))

	defer conn.Close()
	for {
		select {
		case <-exitChan:
			s <- c
			return
		default:
			_, msg, err2 := conn.ReadMessage()
			if err2 != nil {
				if websocket.IsUnexpectedCloseError(err2, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
					logger.Debug(fmt.Sprintf("error parsing message: %v", err2))
				} else {
					logger.Debug(fmt.Sprintf("disconnection error: %v", err2))
				}
				s <- c
				return
			}
			logger.Debug(fmt.Sprintf("from client %d: %s\n", c, msg))
		}
	}
}

func main() {
	var (
		connectionsNumberArg = flag.Int("n", 1, "determine the number of parallel connections")
		verboseArg           = flag.Bool("v", false, "log more information e.g. messages received")
	)

	flag.Parse()

	if *connectionsNumberArg <= 0 {
		log.Fatal("Number of connections must be a positive number")
	}
	var logHandlerOptions *slog.HandlerOptions
	if *verboseArg {
		logHandlerOptions = &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelDebug,
		}
	} else {
		logHandlerOptions = &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelError,
		}
	}

	f, err := os.OpenFile("./"+LOG_FILE, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("could not open log file")
	}
	logger = slog.New(slog.NewTextHandler(f, logHandlerOptions))
	log.Printf("logs in: %s\n", LOG_FILE)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	syncChan := make(chan int, *connectionsNumberArg)
	exitChan := make(chan struct{})
	for i := 0; i < *connectionsNumberArg; i++ {
		go connect(i, syncChan, exitChan)
	}

	counter := 0
	for {
		select {
		case clientIdx := <-syncChan:
			logger.Debug(fmt.Sprintf("client %d was disconnected", clientIdx))
			counter++
		case <-sigChan:
			close(exitChan)
		}
		if counter == *connectionsNumberArg {
			log.Printf("all clients have been disconnected successfully")
			break
		} else {
			log.Printf("disconnected clients: %d", counter)
		}
	}
	log.Println("done :-)")
}
