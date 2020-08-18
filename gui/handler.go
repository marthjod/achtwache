package gui

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	index []byte
}

func NewHandler(index []byte) *Handler {
	return &Handler{
		index: index,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
	if _, err := w.Write(h.index); err != nil {
		log.Error().Err(err).Msg("writing index.html")
	}
}
