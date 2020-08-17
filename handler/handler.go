package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/marthjod/achtwache/client"
	"github.com/marthjod/achtwache/model"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	fieldSelectorKey   = "spec.nodeName"
	concurrentAPICalls = 4
)

type cache struct {
	sync.RWMutex
	contents    []*model.Node
	sema        chan struct{}
	updating    bool
	added       time.Time
	expireAfter time.Duration
	client      *client.Client
}

func (c *cache) setUpdatingFlag(updating bool) {
	c.Lock()
	defer c.Unlock()
	c.updating = updating
}

func (c *cache) Add(contents []*model.Node) {
	c.contents = contents
	c.added = time.Now()
}

func (c *cache) Get(ctx context.Context) []*model.Node {
	if len(c.contents) == 0 {
		c.setUpdatingFlag(true)
		defer c.setUpdatingFlag(false)

		nodes, err := c.update(ctx)
		if err != nil {
			return nil
		}
		c.Add(nodes)
		return c.contents
	}

	if c.added.Add(c.expireAfter).Before(time.Now()) && !c.updating {
		log.Debug().Msgf("cache empty/expired at %s, need refetch", c.added.Add(c.expireAfter))
		c.setUpdatingFlag(true)
		go func() {
			defer c.setUpdatingFlag(false)

			// TODO: context.Background
			nodes, err := c.update(context.Background())
			if err != nil {
				log.Debug().Err(err).Msg("refreshing cache contents")
				return
			}
			c.Add(nodes)
		}()
	}

	return c.contents
}

func (c *cache) update(ctx context.Context) ([]*model.Node, error) {
	var (
		nodes []*model.Node
		wg    sync.WaitGroup
	)
	k8sNodes, _ := c.client.Clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	wg.Add(len(k8sNodes.Items))
	for _, node := range k8sNodes.Items {
		n := &model.Node{}
		n.FromK8s(node)
		nodes = append(nodes, n)

		c.sema <- struct{}{}
		go func(node *model.Node) {
			defer func() {
				<-c.sema
				wg.Done()
			}()

			log.Debug().Msgf("fetching pods for node %s", node.Name)
			k8sPods, _ := c.client.Clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{
				FieldSelector: fieldSelectorKey + "=" + node.Name,
			})
			n.AddPods(k8sPods.Items)
		}(n)
	}
	wg.Wait()
	return nodes, nil
}

type Handler struct {
	cache              *cache
	client             *client.Client
	concurrentAPICalls int
	sem                chan struct{}
}

func New(client *client.Client) *Handler {
	return &Handler{
		cache: &cache{
			expireAfter: 30 * time.Second,
			client:      client,
			sema:        make(chan struct{}, concurrentAPICalls),
		},
	}
}

func (h *Handler) Get(ctx context.Context) ([]*model.Node, error) {
	nodes := h.cache.Get(ctx)
	return nodes, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
