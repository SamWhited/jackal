/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"

	"github.com/stretchr/testify/require"
)

func TestServer_MethodNotSupported(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/route", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	srv := newServer(defaultClusterPort)

	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestServer_BadRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/route", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	srv := newServer(defaultClusterPort)

	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestServer_NotFound(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/unknown_path", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	srv := newServer(defaultClusterPort)

	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestServer_Ok(t *testing.T) {
	j, _ := jid.NewWithString("ortuman@jackal.im/yard", true)
	p := xmpp.NewPresence(j, j.ToBareJID(), xmpp.AvailableType)

	buf := bytes.NewBuffer(nil)
	_ = p.ToXML(buf, true)

	req, err := http.NewRequest(http.MethodPost, "/route", buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", xmlAppMimeType)

	rr := httptest.NewRecorder()
	srv := newServer(defaultClusterPort)

	ch := make(chan xmpp.Stanza, 1)
	srv.registerStanzaHandler(func(_ context.Context, stanza xmpp.Stanza) error {
		ch <- stanza
		return nil
	})

	srv.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, ch, 1)
}
