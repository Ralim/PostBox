package webserver

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (server *WebServer) handleFile(w http.ResponseWriter, r *http.Request) {
	fileRequestedName := r.URL.Path
	file, filename, err := server.fileCache.GetFile(strings.Replace(fileRequestedName, "/", "", -1))
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}
	defer file.Close()
	//Otherwise write out the file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	io.Copy(w, file)
}
