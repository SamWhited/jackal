/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"context"
	"net/http"
	"sync/atomic"

	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/xmpp"
)

const (
	contentTypeHeader = "Content-Type"

	xmlAppMimeType  = "application/xml"
	xmlTextMimeType = "text/xml"

	routePath = "/route"
)

var clusterListerAddr = ":14369"

type StanzaHandler = func(ctx context.Context, stanza xmpp.Stanza) error

type server struct {
	stanzaHnd atomic.Value
	started   int32
	srv       *http.Server
}

func newServer() *server {
	s := &server{}
	s.srv = &http.Server{
		Addr:    clusterListerAddr,
		Handler: s,
	}
	return s
}

func (s *server) start() error {
	if !atomic.CompareAndSwapInt32(&s.started, 0, 1) {
		return nil
	}
	return s.srv.ListenAndServe()
}

func (s *server) registerStanzaHandler(hnd StanzaHandler) {
	s.stanzaHnd.Store(hnd)
}

func (s *server) shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&s.started, 1, 0) {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != routePath {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	contentType := r.Header.Get(contentTypeHeader)
	if contentType != xmlAppMimeType && contentType != xmlTextMimeType {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p := xmpp.NewParser(r.Body, xmpp.DefaultMode, 0)
	elem, err := p.ParseElement()
	if err != nil {
		log.Warnf("failed to parse cluster element: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	stanza, err := xmpp.NewStanzaFromElement(elem)
	if err != nil {
		log.Warnf("invalid stanza: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hnd, ok := s.stanzaHnd.Load().(StanzaHandler)
	if ok {
		_ = hnd(r.Context(), stanza)
	}
	w.WriteHeader(http.StatusOK)
}
