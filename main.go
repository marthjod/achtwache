package main

import (
	"context"
	"os"

	"github.com/marthjod/achtwache/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api/v1/resource"
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
	pods, _ := client.Clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{})

	for _, pod := range pods.Items {
		log.Debug().Msgf("%s", pod.Name)
		requests, limits := resource.PodRequestsAndLimits(&pod)
		log.Debug().Msgf("requests: %v, limits: %v", requests, limits)
	}
}
