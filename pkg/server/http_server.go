package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/config"
	tlsprovider "github.com/protomesh/protomesh/provider/tls"
)

type HttpServer[Dependency any] struct {
	*protomesh.Injector[any]

	Server *http.Server

	TlsBuilder *tlsprovider.TlsBuilder[any] `config:"tls"`

	HttpHandler http.Handler
	GrpcHandler http.Handler

	ShutdownTimeout protomesh.Config `config:"shutdown.timeout,duration" default:"120s" usage:"HTTP server shutdown timeout before closing"`

	closeCh chan error
	addr    string
}

func (h *HttpServer[Dependency]) AssertBeforeStart() error {

	if h.TlsBuilder == nil {
		return errors.New("TlsBuilder is a mandatory attribute in the HttpServer")
	}

	if h.closeCh != nil {
		return errors.New("closeCh must be new (are you trying to reuse a HttpServer?)")
	}

	if h.ShutdownTimeout == nil {
		h.ShutdownTimeout = config.NewConfig(30 * time.Second)
	}

	h.closeCh = make(chan error)

	return nil

}

func (h *HttpServer[Dependency]) Start() {

	h.AssertBeforeStart()

	log := h.Log()

	listener := h.TlsBuilder.BuildListener()

	h.addr = listener.Addr().String()
	h.Server = &http.Server{Handler: h}

	go func() {

		err := h.Server.Serve(listener)

		switch err {

		case http.ErrServerClosed:
			log.Info("Closed http server", "address", h.addr)

		case nil:
			break

		default:
			log.Error("Error closing http server", "address", h.addr, "error", err)
		}

		h.closeCh <- err

	}()

}

func (h *HttpServer[Dependency]) Stop() {

	log := h.Log()

	ctx, cancel := context.WithTimeout(context.TODO(), h.ShutdownTimeout.DurationVal())
	defer cancel()

	err := h.Server.Shutdown(ctx)
	if err != nil {

		log.Error("Http server shutdown error", "address", h.addr, "error", err)

		err = h.Server.Close()
		if err != nil {
			log.Error("Http server close error", "address", h.addr, "error", err)
		}

	}

	<-h.closeCh

}

func (h *HttpServer[Dependency]) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	contentType := r.Header.Get("Content-Type")

	if h.HttpHandler != nil && strings.Contains(contentType, "application/grpc") {
		h.GrpcHandler.ServeHTTP(w, r)
	} else {
		h.HttpHandler.ServeHTTP(w, r)
	}

}
