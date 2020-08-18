package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/marthjod/achtwache/client"
	"github.com/marthjod/achtwache/gui"
	"github.com/marthjod/achtwache/nodes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var (
		namespace  = os.Getenv("NAMESPACE")
		kubeConfig = os.Getenv("KUBE_CONFIG")
		logLevel   = os.Getenv("LOGLEVEL")
		listenAddr = os.Getenv("LISTEN_ADDR")
		indexHTML  = os.Getenv("INDEX_HTML")

		httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of all HTTP requests",
		}, []string{"code", "method"})
	)

	promRegistry := prometheus.NewRegistry()
	promRegistry.MustRegister(httpRequestDuration)

	// TODO
	if indexHTML == "" {
		indexHTML = "index.html"
	}
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

	index, err := ioutil.ReadFile(indexHTML)
	if err != nil {
		log.Error().Err(err).Msgf("reading %s", indexHTML)
	}

	http.Handle("/", gui.NewHandler(index))
	http.Handle("/nodes", promhttp.InstrumentHandlerDuration(httpRequestDuration, nodes.NewHandler(client)))
	http.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}))
	log.Info().Msgf("listening on %s", listenAddr)
	log.Fatal().Err(http.ListenAndServe(listenAddr, nil)).Msg("")
}
