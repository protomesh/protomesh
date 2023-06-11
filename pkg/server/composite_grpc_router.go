package server

import "context"

type CompositeGrpcRouter []GrpcRouter

func NewCompositeGrpcRouter(routers ...GrpcRouter) CompositeGrpcRouter {

	compRouter := []GrpcRouter{}

	compRouter = append(compRouter, routers...)

	return compRouter

}

func (c CompositeGrpcRouter) GetMethod(ctx context.Context, callInfo *GrpcCallInformation) GrpcMethodHandler {

	for _, r := range c {

		m := r.GetMethod(ctx, callInfo)
		if m != nil {
			return m
		}

	}

	return nil

}
