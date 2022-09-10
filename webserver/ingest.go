package webserver

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

// Handle incoming POST or PUT requests
// And decode these into a reader + a filname

func (server WebServer) handlePOST(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(4 * 1024 * 1024) //Max in ram limit
	filesAdded := make([]string, 0, 4)
	if r.MultipartForm != nil {
		for _, fileList := range r.MultipartForm.File {
			for i, header := range fileList {
				file, err := header.Open()
				if err != nil {
					log.Error().Err(err).Str("Filename", header.Filename).Int64("size", header.Size).Int("index", i).Msg("Multi-part form file open failed")
				} else {
					defer file.Close()
					log.Info().Str("Filename", header.Filename).Int64("size", header.Size).Int("index", i).Msg("Multi-part form file")
					server.fileCache.IngestFile(header.Filename, file)
					filesAdded = append(filesAdded, header.Filename)
				}
			}
		}
	}
	log.Info().Strs("filesAdded", filesAdded).Msg("Import done")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
}
