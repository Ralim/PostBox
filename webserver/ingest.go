package webserver

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// Handle incoming POST or PUT requests
// And decode these into a reader + a filname

func (server WebServer) handlePOST(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 * 1024 * 1024); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
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
					if err := server.fileCache.IngestFile(header.Filename, file); err == nil {
						filesAdded = append(filesAdded, header.Filename)
					} else {
						w.WriteHeader(http.StatusInternalServerError)
					}
				}
			}
		}
	}
	log.Info().Strs("filesAdded", filesAdded).Msg("Import done")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("OK"))
}
