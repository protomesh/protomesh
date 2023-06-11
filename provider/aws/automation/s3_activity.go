package automation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/upper-institute/graviflow/internal/config"
	typesaws "github.com/upper-institute/graviflow/proto/api/types/aws"
	typesv1 "github.com/upper-institute/graviflow/proto/api/types/v1"
	"go.temporal.io/sdk/temporal"
)

const (
	s3Err_ApiRequestError      = "ApiRequestError"
	s3Err_InvalidArgumentError = "InvalidArgumentError"
)

func (as *AutomationSet[D]) S3ScanResources(ctx context.Context, req *typesaws.S3ScanResourcesRequest) (*typesaws.S3ScanResourcesResponse, error) {

	s3Cli := as.Dependency().GetAwsS3Client()

	listObjectsReq := &s3.ListObjectsV2Input{
		Bucket: aws.String(req.BucketName),
		Prefix: aws.String(req.KeyPrefix),
	}

	if len(req.ContinuationToken) > 0 {
		listObjectsReq.ContinuationToken = aws.String(req.ContinuationToken)
	}

	listObjectsRes, err := s3Cli.ListObjectsV2(ctx, listObjectsReq)
	if err != nil {
		return nil, temporal.NewApplicationErrorWithCause(
			"Failed to list objects using S3 ListObjectsV2",
			s3Err_ApiRequestError, err,
		)
	}

	res := &typesaws.S3ScanResourcesResponse{
		ResourceObjects: []string{},
	}

	if listObjectsRes.ContinuationToken != nil {
		res.ContinuationToken = aws.ToString(listObjectsRes.ContinuationToken)
	}

	for _, object := range listObjectsRes.Contents {

		lastModified := aws.ToTime(object.LastModified)

		if req.ModifiedSince != nil && req.ModifiedSince.AsTime().Before(lastModified) {
			continue
		}

		res.ResourceObjects = append(res.ResourceObjects, aws.ToString(object.Key))

	}

	return res, nil

}

func (as *AutomationSet[D]) S3ReadFile(ctx context.Context, req *typesaws.S3ReadFileRequest) (*typesaws.S3ReadFileResponse, error) {

	s3Cli := as.Dependency().GetAwsS3Client()

	headRes, err := s3Cli.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(req.BucketName),
		Key:    aws.String(req.ObjectKey),
	})
	if err != nil {
		return nil, temporal.NewApplicationErrorWithCause(
			"Failed to head object using S3 HeadObject",
			s3Err_ApiRequestError, err,
		)
	}

	buf := make([]byte, headRes.ContentLength)

	bufWriter := manager.NewWriteAtBuffer(buf)

	donwloader := manager.NewDownloader(s3Cli, func(d *manager.Downloader) {
		d.Concurrency = 1
	})

	if _, err := donwloader.Download(ctx, bufWriter, &s3.GetObjectInput{
		Bucket: aws.String(req.BucketName),
		Key:    aws.String(req.ObjectKey),
	}); err != nil {
		return nil, temporal.NewApplicationErrorWithCause(
			"Failed to download object from S3",
			s3Err_ApiRequestError, err,
		)
	}

	return &typesaws.S3ReadFileResponse{
		Content: buf,
	}, nil

}

func (as *AutomationSet[D]) S3ObjectToResource(ctx context.Context, req *typesaws.S3ObjectToResourceRequest) (*typesaws.S3ObjectToResourceResponse, error) {

	readFileRes, err := as.S3ReadFile(ctx, &typesaws.S3ReadFileRequest{
		BucketName: req.BucketName,
		ObjectKey:  req.ObjectKey,
	})
	if err != nil {
		return nil, err
	}

	sourceFormat, err := config.ProtoJsonFileExtensionToFormat(req.ObjectKey)
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"Failed to detect protojson source file format",
			s3Err_InvalidArgumentError, err,
		)
	}

	typedRes := &typesv1.TypedResource{}

	if err := config.ProtoJsonUnmarshal(readFileRes.Content, sourceFormat, typedRes); err != nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"Failed to unmarshal file content to proto type",
			s3Err_InvalidArgumentError, err,
		)
	}

	if typedRes.Spec == nil {
		return nil, temporal.NewNonRetryableApplicationError(
			"Empty spec (unknown resource type)",
			s3Err_InvalidArgumentError, err,
		)
	}

	return &typesaws.S3ObjectToResourceResponse{
		Resource: typedRes,
	}, nil

}
