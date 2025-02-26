package httpsrv

import (
	"goapp/internal/pkg/watcher"

	"github.com/gorilla/websocket"
)

func (s *Server) addWatcher(w *watcher.Watcher, c *websocket.Conn) {
	s.watchersLock.Lock()
	defer s.watchersLock.Unlock()
	s.watchers[c] = w
	s.connToUid[c] = w.GetWatcherId()
}

func (s *Server) removeWatcher(w *watcher.Watcher, c *websocket.Conn) {
	s.watchersLock.Lock()
	defer s.watchersLock.Unlock()
	// Print satistics before removing watcher.
	for i := range s.sessionStats {
		if s.sessionStats[i].id == w.GetWatcherId() {
			s.sessionStats[i].print()
		}
	}
	// Remove watcher.
	delete(s.watchers, c)
}

func (s *Server) notifyWatchers(str string) {
	s.watchersLock.RLock()
	defer s.watchersLock.RUnlock()

	// Send message to all watchers and increment stats.
	for wc := range s.watchers {
		s.watchers[wc].Send(str)
		s.incStats(s.connToUid[wc])
	}
}
