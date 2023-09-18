package server

import (
	"errors"
	"io"

	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcHandlerFromGateway(g GrpcGateway, log app.Logger) grpc.StreamHandler {

	return func(srv interface{}, stream grpc.ServerStream) error {

		call, err := g.MatchGrpc(stream)
		if err != nil {
			log.Error("Error matching gRPC call", "error", err)
			return err
		}

		for _, handler := range call.Handlers {

			receiveCh := ReceiveFrom(stream, handler)
			sendCh := SendFrom(stream, handler)

		events:
			for i := 0; i < 4; i++ {

				select {

				case receiveErr, ok := <-receiveCh:
					if !ok {
						receiveCh = nil
						continue events
					}
					if receiveErr != nil {
						log.Debug("Error receiving from stream", "error", receiveErr)
						return receiveErr
					}

				case sendErr, ok := <-sendCh:
					if !ok {
						sendCh = nil
						continue events
					}
					if sendErr != nil {
						log.Debug("Error sending to stream", "error", sendErr)
						return sendErr
					}

				}

			}

		}

		return nil

	}

}

func SendFrom(serverStream grpc.ServerStream, handler gateway.GrpcHandler) <-chan error {

	errCh := make(chan error)

	go func() {

		defer close(errCh)

		errCh <- func() error {
			f := newFrame(nil)

			for i := 0; ; i++ {

				payload, resErr := handler.Result()
				if resErr != nil && resErr != io.EOF {
					_, ok := status.FromError(resErr)
					if ok {
						return resErr
					}

					return status.Errorf(codes.Internal, "Internal error: %s", resErr.Error())

				}

				if i == 0 {
					if md := handler.GetOutgoingMetadata(); md != nil {
						if err := serverStream.SetHeader(md); err != nil {
							return status.Errorf(codes.Internal, "Unable to set header on server stream: %s", err.Error())
						}
					}
				}

				f.payload = payload

				if err := serverStream.SendMsg(f); err != nil {
					return status.Errorf(codes.Internal, "Error sending result to server stream: %s", err.Error())
				}
				if resErr == io.EOF {
					break
				}

			}

			return nil

		}()

	}()

	return errCh

}

func ReceiveFrom(serverStream grpc.ServerStream, handler gateway.GrpcHandler) <-chan error {

	errCh := make(chan error)

	go func() {

		defer close(errCh)

		errCh <- func() error {

			f := newFrame(nil)

			for {

				if err := serverStream.RecvMsg(f); err != nil {

					if errors.Is(err, io.EOF) {
						break
					}

					return err

				}

				err := handler.Call(f.payload)
				if err != nil {

					if errors.Is(err, io.EOF) {
						break
					}

					_, ok := status.FromError(err)
					if ok {
						return err
					}

					return status.Errorf(codes.Internal, "Internal error: %s", err.Error())

				}

			}

			return nil

		}()

	}()

	return errCh

}
