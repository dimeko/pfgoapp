package httpsrv

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type Route struct {
	Name    string
	Method  string
	Pattern string
	HFunc   http.Handler
	Queries []string
}

func (s *Server) myRoutes() []Route {
	routes := []Route{
		{
			Name:    "health",
			Method:  "GET",
			Pattern: "/goapp/health",
			HFunc:   s.handlerWrapper(s.handlerHealth),
		},

		{
			Name:    "home",
			Method:  "GET",
			Pattern: "/goapp",
			HFunc:   s.handlerWrapper(s.handlerHome),
		},
	}

	wsRoute := Route{
		Name:    "websocket",
		Method:  "GET",
		Pattern: "/goapp/ws/{csrfToken}",
		HFunc:   s.handlerWrapper(s.handlerWebSocket),
	}
	if !s.csrfProtection {
		wsRoute.Pattern = "/goapp/ws"
	}
	return append(routes, wsRoute)
}

func (s *Server) handlerWrapper(handlerFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				s.error(w, http.StatusInternalServerError, fmt.Errorf("%v\n%v", r, string(debug.Stack())))
			}
		}()
		handlerFunc(w, r)
	})
}
