package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/upper-institute/graviflow"
	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
	apiv1 "github.com/upper-institute/graviflow/proto/api/v1"
	"github.com/upper-institute/graviflow/provider/postgres/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type resourceStatus string

const (
	activeResource  resourceStatus = "ACTIVE"
	droppedResource resourceStatus = "DROPPED"
)

type ResourceStoreDependency interface {
	GetSqlDatabase() *sql.DB
	GetGrpcServer() *grpc.Server
}

type ResourceStore[D ResourceStoreDependency] struct {
	*graviflow.AppInjector[D]

	apiv1.UnimplementedResourceStoreServer

	MigrationFile graviflow.Config `config:"migration.file,str" usage:"Migration file path to execute"`

	queries *gen.Queries
}

func (r *ResourceStore[D]) Initialize() {

	log := r.Log()

	db := r.Dependency().GetSqlDatabase()

	migrationFile := r.MigrationFile.StringVal()

	if len(migrationFile) > 0 {

		migrations := &migrate.FileMigrationSource{
			Dir: migrationFile,
		}

		appliedMigs, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
		if err != nil {
			log.Panic("Error applying migrations to postgres database", "error", err)
		}

		log.Info("Applied migrations to postgres database", "count", appliedMigs)

	}

	r.queries = gen.New(db)

	apiv1.RegisterResourceStoreServer(r.Dependency().GetGrpcServer(), r)

	log.Info("Postgres ResourceStore registered on gRPC server")

}

func hashResource(res *typesv1.Resource) string {

	hash := sha256.New()

	raw, err := proto.Marshal(res)
	if err != nil {
		panic(err)
	}

	hash.Write(raw)

	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

}

func (r *ResourceStore[D]) Put(ctx context.Context, req *apiv1.PutResourceRequest) (*apiv1.PutResourceResponse, error) {

	id, err := uuid.Parse(req.Resource.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid resource ID, must be in UUID format")
	}

	tx, err := r.Dependency().GetSqlDatabase().Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	versionTimestamp := time.Now()
	versionIndex := versionTimestamp.Unix()

	cache, err := qtx.GetResourceCacheSummary(ctx, gen.GetResourceCacheSummaryParams{
		Namespace: req.Resource.Namespace,
		ID:        id,
	})

	sha256Hash := hashResource(&typesv1.Resource{
		Namespace: req.Resource.Namespace,
		Id:        req.Resource.Id,
		Name:      req.Resource.Name,
		Spec:      req.Resource.Spec,
	})

	if err == sql.ErrNoRows || (err == nil && cache.Sha256Hash != sha256Hash) {

		qtx.UpsertResourceCache(ctx, gen.UpsertResourceCacheParams{
			Namespace:    req.Resource.Namespace,
			ID:           id,
			VersionIndex: versionIndex,
			Name:         req.Resource.Name,
			SpecTypeUrl:  req.Resource.Spec.TypeUrl,
			SpecValue:    req.Resource.Spec.Value,
			Sha256Hash:   sha256Hash,
		})

		qtx.InsertActiveResourceEvent(ctx, gen.InsertActiveResourceEventParams{
			Namespace: req.Resource.Namespace,
			ID:        id,
		})

	} else if err == nil {

		versionIndex = cache.VersionIndex
		versionTimestamp = time.Unix(versionIndex, 0)

	} else if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &apiv1.PutResourceResponse{
		Version: &typesv1.Version{
			Sha256Hash: sha256Hash,
			Timestamp:  timestamppb.New(versionTimestamp),
			Index:      versionIndex,
		},
	}, nil

}

func (r *ResourceStore[D]) Drop(ctx context.Context, req *apiv1.DropResourcesRequest) (*apiv1.DropResourcesResponse, error) {

	tx, err := r.Dependency().GetSqlDatabase().Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	versionTimestamp := time.Now()
	versionIndex := versionTimestamp.Unix()

	for i, idStr := range req.ResourceIds {

		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid resource ID, must be in UUID format (index %d value %s)", i, idStr)
		}

		if err := qtx.InsertDroppedResourceEvent(ctx, gen.InsertDroppedResourceEventParams{
			Namespace:    req.Namespace,
			ID:           id,
			VersionIndex: versionIndex,
		}); err != nil {
			return nil, err
		}

		if err := qtx.DropResourceCache(ctx, gen.DropResourceCacheParams{
			Namespace: req.Namespace,
			ID:        id,
		}); err != nil {
			return nil, err
		}

	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &apiv1.DropResourcesResponse{}, nil

}

func (r *ResourceStore[D]) DropBefore(ctx context.Context, req *apiv1.DropBeforeResourcesRequest) (*apiv1.DropBeforeResourcesResponse, error) {

	tx, err := r.Dependency().GetSqlDatabase().Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	count, err := qtx.CountResourceCacheBefore(ctx, gen.CountResourceCacheBeforeParams{
		Namespace:          req.Namespace,
		BeforeVersionIndex: req.Before,
	})
	if err != nil {
		return nil, err
	}

	versionTimestamp := time.Now()
	versionIndex := versionTimestamp.Unix()

	if err := qtx.InsertDroppedResourceBeforeEvent(ctx, gen.InsertDroppedResourceBeforeEventParams{
		Namespace:          req.Namespace,
		BeforeVersionIndex: req.Before,
		VersionIndex:       versionIndex,
	}); err != nil {
		return nil, err
	}

	if err := qtx.DropResourceCacheBefore(ctx, gen.DropResourceCacheBeforeParams{
		Namespace:          req.Namespace,
		BeforeVersionIndex: req.Before,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &apiv1.DropBeforeResourcesResponse{
		DroppedCount: int32(count),
	}, nil

}

func (r *ResourceStore[D]) List(stream apiv1.ResourceStore_ListServer) error {

	ctx := stream.Context()

	var listEventsParams *gen.ListResourcesEventsFromNamespaceParams

	for {

		req, err := stream.Recv()
		if err != nil {

			if errors.Is(err, io.EOF) {
				break
			}

			return err

		}

		if listEventsParams == nil {

			pageSize := int32(50)

			for i := int32(0); ; i++ {

				rows, err := r.queries.ListResourcesCachedFromNamespace(ctx, gen.ListResourcesCachedFromNamespaceParams{
					Namespace:  req.Namespace,
					MaxRows:    pageSize,
					RowsOffset: pageSize * i,
				})
				if err != nil {
					return err
				}

				res := &apiv1.ListResourcesResponse{
					UpdatedResources: []*typesv1.Resource{},
					DroppedResources: []*typesv1.Resource{},
					Nonce:            req.Nonce,
					EndOfList:        false,
				}

				for _, row := range rows {

					res.UpdatedResources = append(res.UpdatedResources, &typesv1.Resource{
						Namespace: req.Namespace,
						Id:        row.ID.String(),
						Name:      row.Name,
						Spec: &anypb.Any{
							TypeUrl: row.SpecTypeUrl,
							Value:   row.SpecValue,
						},
						Version: &typesv1.Version{
							Sha256Hash: row.Sha256Hash,
							Index:      row.VersionIndex,
							Timestamp:  timestamppb.New(time.Unix(row.VersionIndex, 0)),
						},
					})

				}

				if len(rows) > 0 {

					if err := stream.Send(res); err != nil {
						return err
					}

				}

				if len(rows) < int(pageSize) {
					break
				}

			}

			maxIndex, err := r.queries.MaxVersionIndexForNamespace(ctx, req.Namespace)
			if err != nil {
				return err
			}

			listEventsParams = &gen.ListResourcesEventsFromNamespaceParams{
				Namespace:        req.Namespace,
				MaxRows:          pageSize,
				FromVersionIndex: maxIndex.VersionIndex,
				FromID:           maxIndex.ID,
			}

			if err := stream.Send(&apiv1.ListResourcesResponse{
				Nonce:     req.Nonce,
				EndOfList: true,
			}); err != nil {
				return err
			}

			continue

		}

		for {

			rows, err := r.queries.ListResourcesEventsFromNamespace(ctx, *listEventsParams)
			if err != nil {
				return err
			}

			res := &apiv1.ListResourcesResponse{
				UpdatedResources: []*typesv1.Resource{},
				DroppedResources: []*typesv1.Resource{},
				Nonce:            req.Nonce,
				EndOfList:        false,
			}

			for _, row := range rows {

				rowResource := &typesv1.Resource{
					Namespace: req.Namespace,
					Id:        row.ID.String(),
					Name:      row.Name,
					Spec: &anypb.Any{
						TypeUrl: row.SpecTypeUrl,
						Value:   row.SpecValue,
					},
					Version: &typesv1.Version{
						Sha256Hash: row.Sha256Hash,
						Index:      row.VersionIndex,
						Timestamp:  timestamppb.New(time.Unix(row.VersionIndex, 0)),
					},
				}

				switch resourceStatus(row.Status) {

				case activeResource:
					res.UpdatedResources = append(res.UpdatedResources, rowResource)

				case droppedResource:
					res.DroppedResources = append(res.DroppedResources, rowResource)

				}

				listEventsParams.FromVersionIndex = row.VersionIndex
				listEventsParams.FromID = row.ID

			}

			if len(rows) > 0 {

				if err := stream.Send(res); err != nil {
					return err
				}

			}

			if len(rows) < int(listEventsParams.MaxRows) {
				break
			}

		}

		if err := stream.Send(&apiv1.ListResourcesResponse{
			Nonce:     req.Nonce,
			EndOfList: true,
		}); err != nil {
			return err
		}

	}

	return nil

}
