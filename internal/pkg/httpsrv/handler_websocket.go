package httpsrv

import (
	"fmt"
	"goapp/internal/pkg/watcher"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func (s *Server) handlerWebSocket(w http.ResponseWriter, r *http.Request) {
	if s.csrfProtection {
		// Get csrfToken
		csrfToken := mux.Vars(r)["csrfToken"]

		if !s.isValidCsrfToken(csrfToken) {
			s.error(w, http.StatusInternalServerError, fmt.Errorf("bad csrf token"))
			return
		}
		// Delete csrfToken now in order to prevent new connections
		s.deleteCsrfToken(csrfToken)
	}

	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if err := s.epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection")
		conn.Close()
	}
	var watch = watcher.New()
	if err := watch.Start(); err != nil {
		s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to start watcher: %w", err))
		return
	}
	s.addWatcher(watch, conn)
}
