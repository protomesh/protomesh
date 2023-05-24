package server

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GrpcMethodHandler interface {
	Call([]byte) (metadata.MD, error)
	Result() ([]byte, error)
}

type GrpcMethod interface {
	Handle(context.Context) (GrpcMethodHandler, error)
}

type GrpcRouter interface {
	GetMethod(string) GrpcMethod
}

type GrpcProxy struct {
	Router GrpcRouter
}

func (gp *GrpcProxy) Handle(srv interface{}, serverStream grpc.ServerStream) error {

	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Unimplemented, "%s does not exists", fullMethodName)
	}

	method := gp.Router.GetMethod(fullMethodName)
	if method == nil {
		return status.Errorf(codes.Unimplemented, "%s does not have handler", fullMethodName)
	}

	ctx := serverStream.Context()

	handler, err := method.Handle(ctx)
	if err != nil {

		_, ok := status.FromError(err)
		if ok {
			return err
		}

		return status.Errorf(codes.Internal, "Internal error: %s", err.Error())

	}

	receiveCh := ReceiveFrom(serverStream, handler)
	sendCh := SendFrom(serverStream, handler)

	for i := 0; i < 2; i++ {

		select {

		case receiveErr := <-receiveCh:
			if receiveErr != nil {
				return receiveErr
			}

		case sendErr := <-sendCh:
			if sendErr != nil {
				return sendErr
			}

		}

	}

	return nil

}

func SendFrom(serverStream grpc.ServerStream, handler GrpcMethodHandler) <-chan error {

	errCh := make(chan error)

	go func() {

		defer close(errCh)

		errCh <- func() error {
			f := NewFrame(nil)

			for {

				payload, err := handler.Result()
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

				f.payload = payload

				if err := serverStream.SendMsg(f); err != nil {
					return status.Errorf(codes.Internal, "Error sending result to server stream: %s", err.Error())
				}

			}

			return nil

		}()

	}()

	return errCh

}

func ReceiveFrom(serverStream grpc.ServerStream, handler GrpcMethodHandler) <-chan error {

	errCh := make(chan error)

	go func() {

		defer close(errCh)

		errCh <- func() error {

			f := NewFrame(nil)

			for i := 0; ; i++ {

				if err := serverStream.RecvMsg(f); err != nil {

					if errors.Is(err, io.EOF) {
						break
					}

					return err

				}

				md, err := handler.Call(f.payload)
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

				if i == 0 {
					if err := serverStream.SetHeader(md); err != nil {
						return status.Errorf(codes.Internal, "Unable to set header on server stream: %s", err.Error())
					}
				}

			}

			return nil

		}()

	}()

	return errCh

}
