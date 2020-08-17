package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/marthjod/achtwache/client"
	"github.com/marthjod/achtwache/model"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	fieldSelectorKey   = "spec.nodeName"
	concurrentAPICalls = 4
)

type Handler struct {
	client             *client.Client
	concurrentAPICalls int
	sem                chan struct{}
	// TODO: cache
}

func New(client *client.Client) *Handler {
	return &Handler{
		client:             client,
		concurrentAPICalls: concurrentAPICalls,
		sem:                make(chan struct{}, concurrentAPICalls),
	}
}

func (h *Handler) Update(ctx context.Context) ([]*model.Node, error) {
	var (
		nodes []*model.Node
		wg    sync.WaitGroup
	)
	k8sNodes, _ := h.client.Clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	wg.Add(len(k8sNodes.Items))
	for _, node := range k8sNodes.Items {
		n := &model.Node{}
		n.FromK8s(node)
		nodes = append(nodes, n)

		h.sem <- struct{}{}
		go func(node *model.Node) {
			defer func() {
				<-h.sem
				wg.Done()
			}()

			log.Debug().Msgf("fetching pods for node %s", node.Name)
			k8sPods, _ := h.client.Clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{
				FieldSelector: fieldSelectorKey + "=" + node.Name,
			})
			n.AddPods(k8sPods.Items)
		}(n)
	}
	wg.Wait()
	return nodes, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.Update(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
