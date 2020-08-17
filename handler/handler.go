package handler

import (
	"context"
	"sync"

	"github.com/marthjod/achtwache/client"
	"github.com/marthjod/achtwache/model"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const fieldSelectorKey = "spec.nodeName"

type Handler struct {
	Client *client.Client
	sem    chan struct{}
	// TODO: cache
}

func (h *Handler) Update(ctx context.Context, concurreny int) ([]*model.Node, error) {
	var (
		nodes []*model.Node
		wg    sync.WaitGroup
		sem   = make(chan struct{}, concurreny)
	)
	k8sNodes, _ := h.Client.Clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	wg.Add(len(k8sNodes.Items))
	for _, node := range k8sNodes.Items {
		n := &model.Node{}
		n.FromK8s(node)
		nodes = append(nodes, n)

		sem <- struct{}{}
		go func(node *model.Node) {
			defer func() {
				wg.Done()
				<-sem
			}()

			log.Debug().Msgf("fetching pods for node %s", node.Name)
			k8sPods, _ := h.Client.Clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{
				FieldSelector: fieldSelectorKey + "=" + node.Name,
			})
			n.AddPods(k8sPods.Items)
		}(n)
	}
	wg.Wait()
	return nodes, nil
}
