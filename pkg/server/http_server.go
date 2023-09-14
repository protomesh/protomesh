package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/config"
	"github.com/protomesh/protomesh/pkg/gateway"
	tlsprovider "github.com/protomesh/protomesh/provider/tls"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type HttpGateway interface {
	MatchHttp(res http.ResponseWriter, req *http.Request) (*gateway.HttpCall, error)
}

type HttpServer[D any] struct {
	*app.Injector[any]

	Server *http.Server

	TlsBuilder *tlsprovider.TlsBuilder[any] `config:"tls"`

	Gateway HttpGateway

	GrpcHandler http.Handler

	ShutdownTimeout app.Config `config:"shutdown.timeout,duration" default:"120s" usage:"HTTP server shutdown timeout before closing"`

	DisableTls app.Config `config:"tls.disable" default:"false" usage:"Disable TLS"`

	closeCh chan error
	addr    string
}

func (h *HttpServer[D]) AssertBeforeStart() error {

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

func (h *HttpServer[D]) Start() {

	h.AssertBeforeStart()

	log := h.Log()

	var listener net.Listener

	if h.DisableTls.BoolVal() {
		rawListener, err := net.Listen(h.TlsBuilder.Protocol.StringVal(), h.TlsBuilder.ListenerAddress.StringVal())

		if err != nil {
			log.Panic("Error creating http listener", "address", h.TlsBuilder.ListenerAddress.StringVal(), "error", err)
			return
		}

		listener = rawListener
	} else {
		listener = h.TlsBuilder.BuildListener()
	}

	h.addr = listener.Addr().String()

	if h.DisableTls.BoolVal() {

		h.Server = &http.Server{
			Handler: h2c.NewHandler(h, &http2.Server{}),
		}

	} else {
		h.Server = &http.Server{Handler: h}
	}

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

func (h *HttpServer[D]) Stop() {

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

func (h *HttpServer[D]) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log := h.Log()

	requestId := r.Header.Get("X-Request-Id")
	if len(requestId) > 0 {
		log = log.With("requestId", requestId)
	}

	contentType := r.Header.Get("Content-Type")

	if h.GrpcHandler != nil && strings.Contains(contentType, "application/grpc") {
		h.GrpcHandler.ServeHTTP(w, r)
		return
	}

	call, err := h.Gateway.MatchHttp(w, r)

	if err == http.ErrNotSupported {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Error matching http call", "error", err)
		return
	}

	for _, handler := range call.Handlers {

		err := handler.Call()
		if err == http.ErrAbortHandler {
			return
		} else if err == io.EOF {
			continue
		} else if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			log.Error("Error calling http handler", "error", err)
			return
		}
	}

}
