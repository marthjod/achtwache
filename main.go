package main

import (
	"net/http"
	"os"

	"github.com/marthjod/achtwache/client"
	"github.com/marthjod/achtwache/handler"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var (
		namespace  = os.Getenv("NAMESPACE")
		kubeConfig = os.Getenv("KUBE_CONFIG")
		logLevel   = os.Getenv("LOGLEVEL")
		listenAddr = os.Getenv("LISTEN_ADDR")
	)

	// TODO
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	if logLevel == "" {
		logLevel = "info"
	}
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Err(err).Msg("parsing log level")
	}
	zerolog.SetGlobalLevel(lvl)

	var opts = []client.Option{
		client.WithNamespace(namespace),
	}

	if kubeConfig == "" {
		log.Info().Msg("defaulting to in-cluster config")
		opts = append(opts, client.WithInClusterConfig())
	} else {
		opts = append(opts, client.WithKubeConfig(kubeConfig))
	}

	client, err := client.New(opts...)
	if err != nil {
		log.Fatal().Err(err).Msg("creating client")
	}

	hdlr := handler.New(client)
	http.Handle("/", hdlr)
	log.Info().Msgf("listening on %s", listenAddr)
	log.Fatal().Err(http.ListenAndServe(listenAddr, nil)).Msg("")
}
