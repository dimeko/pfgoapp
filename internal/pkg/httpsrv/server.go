package httpsrv

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"goapp/internal/pkg/epoll"
	"goapp/internal/pkg/watcher"
	"goapp/pkg/util"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const CSRF_TOKEN_TTL = 3600 // seconds

type Server struct {
	strChan      <-chan string                        // String channel.
	server       *http.Server                         // Gorilla HTTP server.
	watchers     map[*websocket.Conn]*watcher.Watcher // Counter watchers (k: counterId).
	watchersLock *sync.RWMutex                        // Counter lock.
	sessionStats []*sessionStats                      // Session stats.
	quitChannel  chan struct{}                        // Quit channel.
	running      sync.WaitGroup                       // Running goroutines.

	// New fields
	csrfProtection bool
	csrfTokens     map[string]int64
	csrfTokensLock *sync.RWMutex
	epoller        *epoll.Epollt
	connToUid      map[*websocket.Conn]string
}

func New(strChan <-chan string, csrfProtection bool) *Server {
	s := Server{}
	s.strChan = strChan
	s.server = nil // Set below.
	s.watchers = make(map[*websocket.Conn]*watcher.Watcher)
	s.watchersLock = &sync.RWMutex{}
	s.sessionStats = []*sessionStats{}
	s.quitChannel = make(chan struct{})
	s.running = sync.WaitGroup{}

	s.csrfProtection = csrfProtection
	s.csrfTokens = make(map[string]int64)
	s.csrfTokensLock = &sync.RWMutex{}
	epoller, err := epoll.MkEpoll()
	if err != nil {
		panic(err)
	}
	s.epoller = epoller
	s.connToUid = make(map[*websocket.Conn]string)
	return &s
}

func (s *Server) Start() error {
	// Create router.
	r := mux.NewRouter()

	// Register routes.
	for _, route := range s.myRoutes() {
		if route.Method == "ANY" {
			r.Handle(route.Pattern, route.HFunc)
		} else {
			r.Handle(route.Pattern, route.HFunc).Methods(route.Method)
			if route.Queries != nil {
				r.Handle(route.Pattern, route.HFunc).Methods(route.Method).Queries(route.Queries...)
			}
		}
	}

	// Create HTTP server.
	s.server = &http.Server{
		Addr:         "localhost:8080",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  10 * time.Second,
		Handler:      handlers.CombinedLoggingHandler(os.Stdout, r),
	}

	// Start HTTP server.
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	s.running.Add(3)
	go s.mainLoop()
	go s.invalidateCsrfTokens()
	go s.wsMainLoop()

	return nil
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("error: %v\n", err)
	}

	close(s.quitChannel)
	s.running.Wait()
}

func (s *Server) mainLoop() {
	defer s.running.Done()

	for {
		select {
		case str := <-s.strChan:
			s.notifyWatchers(str)
		case <-s.quitChannel:
			return
		}
	}
}

// Loop over the csrfTokens list and delete the expired tokens
func (s *Server) invalidateCsrfTokens() {
	defer s.running.Done()

	for {
		select {
		case <-s.quitChannel:
			return
		default:
			s.csrfTokensLock.Lock()
			for k, v := range s.csrfTokens {
				if util.TTLExpired(v, s.getCsrfTtl()) {
					delete(s.csrfTokens, k)
				}
			}
			s.csrfTokensLock.Unlock()
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) isValidCsrfToken(_csrfToken string) bool {
	s.csrfTokensLock.RLock()
	defer s.csrfTokensLock.RUnlock()
	if createdAt, ok := s.csrfTokens[_csrfToken]; ok {
		if util.TTLExpired(createdAt, s.getCsrfTtl()) {
			return false
		}
	} else {
		return false
	}
	return true
}

func (s *Server) deleteCsrfToken(_csrfToken string) {
	s.csrfTokensLock.Lock()
	defer s.csrfTokensLock.Unlock()
	if _, ok := s.csrfTokens[_csrfToken]; ok {
		delete(s.csrfTokens, _csrfToken)
	}
}

func (s *Server) getCsrfTtl() int64 {
	return CSRF_TOKEN_TTL
}

func (s *Server) wsMainLoop() {
	defer s.running.Done()

	doneCh := make(chan struct{})
	defer close(doneCh)
	for {
		select {
		default:
			connections, err := s.epoller.Wait()
			if err != nil {
				log.Printf("Failed to epoll wait %v", err)
				continue
			}
			time.Sleep(1 * time.Second)
			wg := sync.WaitGroup{}
			for _, conn := range connections {
				wg.Add(1)
				go func(cc *websocket.Conn) {
					defer wg.Done()
					select {
					case cv := <-s.watchers[cc].Recv():
						data, _ := json.Marshal(cv)
						err = cc.WriteMessage(websocket.TextMessage, data)
						if err != nil {
							if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
								log.Printf("failed to write message: %v\n", err)
							}
							return
						}
						return
					}
				}(conn)
			}

			wg.Wait()
		case <-doneCh:
			return
		case <-s.quitChannel:
			return
		}
	}
}
