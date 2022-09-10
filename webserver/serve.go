package webserver

import (
	"fmt"
	"github.com/justinas/alice"
	"github.com/ralim/PostBox/webserver/cache"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"net/http"
	"path"
	"strings"
	"time"
)

type WebServer struct {
	httpServer *http.Server
	fileCache  *cache.FileCache
}

func NewServer() *WebServer {
	return &WebServer{
		httpServer: nil,
		fileCache:  cache.NewFileCache(),
	}
}

func (server *WebServer) StartWebserver() {

	c := alice.New()

	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(log.Logger))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to that handler, all our logs will come with some prepopulated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Debug().
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg(r.Method)
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))

	// Here is your final handleS
	h := c.Then(server)
	server.httpServer = &http.Server{Addr: fmt.Sprintf(":%d", 8080), Handler: h}
	if err := server.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error().Err(err).Msg("HTTP server closed")
	} else {
		log.Warn().Msg("HTTP server closed")
	}
}

func (server *WebServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)
	//Healthcheck does not require auth so chcek it first
	if head == "healthcheck" {
		res.WriteHeader(http.StatusOK)
		return
	}
	if req.Method == http.MethodPost {
		server.handlePOST(res, req)
		return
	}

	switch head {
	//case "file":
	//	server.httpHandlevFile(res, req)
	case "file":
		server.handleFile(res, req)
	case "index.html":
		fallthrough
	case "":
		fallthrough
	case "/":
		server.handleIndex(res, req)
	default:
		res.WriteHeader(http.StatusNotFound)
	}
}

// ShiftPath splits off the front portion of the provided path into head and then returns the remainder in tail
func ShiftPath(pathIn string) (head, tail string) {
	pathIn = path.Clean("/" + pathIn)
	i := strings.Index(pathIn[1:], "/") + 1
	if i <= 0 {
		return pathIn[1:], "/"
	}
	return pathIn[1:i], pathIn[i:]
}
