package main

import (
	"context"
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
	)

	if logLevel == "" {
		logLevel = "info"
	}
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Warn().Err(err).Msg("parsing log level")
	}
	zerolog.SetGlobalLevel(lvl)

	ctx := context.Background()

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

	hdlr := &handler.Handler{
		Client: client,
	}
	nodes, err := hdlr.Update(ctx, 4)
	if err != nil {
		log.Fatal().Err(err).Msg("updating")
	}
	for _, node := range nodes {
		log.Debug().Msgf("node: %s", node.Name)
		for _, pod := range node.Pods {
			log.Debug().Msgf("  %s (%s %s)", pod.Name, pod.CPU.String(), pod.Memory.String())
		}
	}
}
