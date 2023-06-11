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
	Call([]byte) error
	GetResult() ([]byte, error)
	GetOutgoingMetadata() metadata.MD
}

type GrpcCallInformation struct {
	FullMethodName string
}

type GrpcRouter interface {
	GetMethod(context.Context, *GrpcCallInformation) GrpcMethodHandler
}

type GrpcProxy struct {
	Router GrpcRouter
}

func (gp *GrpcProxy) Handle(srv interface{}, serverStream grpc.ServerStream) error {

	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Unimplemented, "%s does not exists", fullMethodName)
	}
	ctx := serverStream.Context()

	callInfo := &GrpcCallInformation{
		FullMethodName: fullMethodName,
	}

	handler := gp.Router.GetMethod(ctx, callInfo)
	if handler == nil {
		return status.Errorf(codes.Unimplemented, "%s does not have handler", fullMethodName)
	}

	// handler, err := method.Handle(ctx)
	// if err != nil {

	// 	_, ok := status.FromError(err)
	// 	if ok {
	// 		return err
	// 	}

	// 	return status.Errorf(codes.Internal, "Internal error: %s", err.Error())

	// }

	receiveCh := ReceiveFrom(serverStream, handler)
	sendCh := SendFrom(serverStream, handler)

events:
	for i := 0; i < 4; i++ {

		select {

		case receiveErr, ok := <-receiveCh:
			if !ok {
				receiveCh = nil
				continue events
			}
			if receiveErr != nil {
				return receiveErr
			}

		case sendErr, ok := <-sendCh:
			if !ok {
				sendCh = nil
				continue events
			}
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
			f := newFrame(nil)

			for i := 0; ; i++ {

				payload, resErr := handler.GetResult()
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

func ReceiveFrom(serverStream grpc.ServerStream, handler GrpcMethodHandler) <-chan error {

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
