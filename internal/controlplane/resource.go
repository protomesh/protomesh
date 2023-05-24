package controlplane

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"time"

	typesv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/types/v1"
	apiv1 "dev.azure.com/pomwm/pom-tech/graviflow/proto/api/v1"
)

func idFromServiceMeshNode(namespace string, node *typesv1.ServiceMesh_Node) string {

	hash := sha256.New()

	switch node := node.Node.(type) {

	case *typesv1.ServiceMesh_Node_HttpIngress:
		hash.Write([]byte(node.HttpIngress.XdsClusterName))
		hash.Write([]byte(node.HttpIngress.IngressName))

	case *typesv1.ServiceMesh_Node_RoutingPolicy:
		hash.Write([]byte(node.RoutingPolicy.XdsClusterName))
		hash.Write([]byte(node.RoutingPolicy.PolicyName))

	case *typesv1.ServiceMesh_Node_Service:
		hash.Write([]byte(node.Service.ServiceName))

	}

	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

}

func logKvsFromResource(res *typesv1.Resource) []interface{} {
	return []interface{}{
		"resourceId", res.Id,
		"resourceName", res.Name,
	}
}

type ResourceStoreSynchronizer struct {
	SyncInterval  time.Duration
	Namespace     string
	ResourceStore apiv1.ResourceStoreClient
	IndexCursor   int64

	EventHandler interface {
		OnBeforeProcess(context.Context) error
		OnUpdated(context.Context, *typesv1.Resource) error
		OnDropped(context.Context, *typesv1.Resource) error
		OnAfterProcess(context.Context) error
	}
}

func (rss *ResourceStoreSynchronizer) Sync(ctx context.Context) <-chan error {

	errCh := make(chan error)

	go func() {

		errCh <- func() error {

			listCli, err := rss.ResourceStore.List(ctx)
			if err != nil {
				return err
			}

			listResCh := make(chan *apiv1.ListResourcesResponse)
			recvErrCh := make(chan error)

			go func() {

				defer close(listResCh)
				defer close(recvErrCh)

				for {

					if ctx.Err() != nil {
						break
					}

					listRes, err := listCli.Recv()
					if errors.Is(err, io.EOF) {
						recvErrCh <- nil
						break
					}

					if err != nil {
						recvErrCh <- err
						break
					}

					listResCh <- listRes

				}

			}()

			for {

				nonce := time.Now().Format(time.RFC3339)

				err := listCli.Send(&apiv1.ListResourcesRequest{
					UpdatedSince: rss.IndexCursor,
					Namespace:    rss.Namespace,
					Nonce:        nonce,
				})
				if err != nil {
					return err
				}

				syncInterval := time.NewTimer(rss.SyncInterval)

			inner:
				for {

					select {

					case listRes := <-listResCh:

						syncInterval.Stop()

						if listRes.Nonce == nonce {
							break inner
						}

						if err := rss.EventHandler.OnBeforeProcess(ctx); err != nil {
							return err
						}

						for _, updatedRes := range listRes.UpdatedResources {

							if updatedRes.Version.Index > rss.IndexCursor {
								rss.IndexCursor = updatedRes.Version.Index
							}

							if err := rss.EventHandler.OnUpdated(ctx, updatedRes); err != nil {
								return err
							}

						}

						for _, droppedRes := range listRes.DroppedResources {

							if droppedRes.Version.Index > rss.IndexCursor {
								rss.IndexCursor = droppedRes.Version.Index
							}

							if err := rss.EventHandler.OnDropped(ctx, droppedRes); err != nil {
								return err
							}

						}

						if err := rss.EventHandler.OnAfterProcess(ctx); err != nil {
							return err
						}

						syncInterval.Reset(rss.SyncInterval)

					case err := <-recvErrCh:
						return err

					case <-syncInterval.C:
						break inner

					case <-ctx.Done():
						return ctx.Err()

					}
				}

			}

		}()

	}()

	return errCh

}
