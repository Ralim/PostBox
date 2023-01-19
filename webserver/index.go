package webserver

import (
	"fmt"
	"net/http"
)

// Renders out a virtual webpage of stored files
//

func (server *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {

	_, _ = w.Write([]byte(fmt.Sprintf("<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 3.2 Final//EN\">\n<html>\n <head>\n  <title>Index of /</title>\n </head>\n"+
		"<body>\n<h1>Index of /</h1>\n"+
		"<h2>Reminder, use curl http://%s -u username:password -F file=@local/file/path</h2>\n"+
		"<ul>", r.Host)))

	allFiles := server.fileCache.ListFiles()
	for _, file := range allFiles {
		_, _ = w.Write([]byte(fmt.Sprintf("<li><a href=\"/file/%s\"> %s/</a></li>\n", file.Key, file.Name)))
	}
	_, _ = w.Write([]byte("</ul>\n</body></html>"))
}
