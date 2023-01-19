package webserver

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/justinas/alice"
	"github.com/ralim/PostBox/webserver/cache"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

type WebServer struct {
	httpServer   *http.Server
	fileCache    *cache.FileCache
	authUserHash []byte
	authPassHash []byte
	authEnabled  bool
}

func NewServer(userHash, passwordHash []byte) *WebServer {
	return &WebServer{
		httpServer:   nil,
		fileCache:    cache.NewFileCache(),
		authUserHash: userHash,
		authPassHash: passwordHash,
		authEnabled:  len(userHash) > 0 && len(passwordHash) > 0,
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

	if server.authEnabled {
		username, password, ok := req.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], server.authUserHash) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], server.authPassHash) == 1)
			if !usernameMatch || !passwordMatch {
				ok = false
			}

		}
		if !ok {
			res.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}
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
