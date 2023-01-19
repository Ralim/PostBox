package main

import (
	"encoding/hex"
	"os"

	"github.com/ralim/PostBox/webserver"
	"github.com/rs/zerolog/log"
)

func main() {
	authUserHashStr := os.Getenv("AUTH_USER_HASH")
	authPassHashStr := os.Getenv("AUTH_PASS_HASH")

	authUserHash, err := hex.DecodeString(authUserHashStr)
	if err != nil {
		log.Error().Msg("Could not decode user auth hash as hex sha256 string")
		panic(err)
	}

	authPassHash, err := hex.DecodeString(authPassHashStr)
	if err != nil {
		log.Error().Msg("Could not decode user pass hash as hex sha256 string")
		panic(err)
	}

	server := webserver.NewServer(authUserHash,
		authPassHash)
	server.StartWebserver()

}
