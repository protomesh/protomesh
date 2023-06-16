package resource

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/protomesh/go-app"
	servicesv1 "github.com/protomesh/protomesh/proto/api/services/v1"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

var (
	WorkflowIdNamespace = uuid.MustParse("3d8e41b4-f7d9-11ed-b67e-0242ac120002")
)

type ResourceStoreSynchronizerDependency interface {
	GetResourceStoreClient() servicesv1.ResourceStoreClient
}

type ResourceStoreSynchronizer[D ResourceStoreSynchronizerDependency] struct {
	*app.Injector[D]

	Namespace   string
	IndexCursor int64

	EventHandler interface {
		BeforeBatch(context.Context) error
		OnUpdated(context.Context, *typesv1.Resource) error
		OnDropped(context.Context, *typesv1.Resource) error
		AfterBatch(context.Context) error
	}
}

func (rss *ResourceStoreSynchronizer[D]) Sync(ctx context.Context) <-chan error {

	log := rss.Log()

	errCh := make(chan error)

	resCli := rss.Dependency().GetResourceStoreClient()

	go func() {

		log.Info("Starting resource sync goroutine")

		err := func() error {

			log.Info("Starting list resource session")

			log.Debug("Sending list resource request", "namespace", rss.Namespace)

			watchStream, err := resCli.Watch(ctx, &servicesv1.WatchResourcesRequest{
				Namespace: rss.Namespace,
			})
			if err != nil {
				return err
			}

			listResCh := make(chan *servicesv1.WatchResourcesResponse)
			recvErrCh := make(chan error)

			go func() {

				defer close(listResCh)
				defer close(recvErrCh)

				for {

					if ctx.Err() != nil {
						break
					}

					listRes, err := watchStream.Recv()
					if errors.Is(err, io.EOF) {
						recvErrCh <- nil
						break
					}

					if err != nil {
						recvErrCh <- err
						break
					}

					log.Debug(
						"Received resource list",
						"updatedResources", len(listRes.UpdatedResources),
						"droppedResources", len(listRes.DroppedResources),
					)

					listResCh <- listRes

				}

			}()

		inner:
			for {

				select {

				case listRes := <-listResCh:

					if listRes.EndOfList {
						continue inner
					}

					if err := rss.EventHandler.BeforeBatch(ctx); err != nil {
						return err
					}

					for _, updatedRes := range listRes.UpdatedResources {

						log.Debug("Updated resource", "id", updatedRes.Id, "name", updatedRes.Name)

						if updatedRes.Version.Index > rss.IndexCursor {
							rss.IndexCursor = updatedRes.Version.Index
						}

						if err := rss.EventHandler.OnUpdated(ctx, updatedRes); err != nil {
							return err
						}

					}

					for _, droppedRes := range listRes.DroppedResources {

						log.Debug("Dropped resource", "id", droppedRes.Id, "name", droppedRes.Name)

						if droppedRes.Version.Index > rss.IndexCursor {
							rss.IndexCursor = droppedRes.Version.Index
						}

						if err := rss.EventHandler.OnDropped(ctx, droppedRes); err != nil {
							return err
						}

					}

					if err := rss.EventHandler.AfterBatch(ctx); err != nil {
						return err
					}

				case err := <-recvErrCh:
					log.Error("Error waiting for resource list response", "error", err)
					return err

				case <-ctx.Done():
					log.Debug("Sync context done")
					return ctx.Err()

				}

			}

		}()

		log.Panic("Sync resource routine error", "error", err)

		errCh <- err

	}()

	return errCh

}
