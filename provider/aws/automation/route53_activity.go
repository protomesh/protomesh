package automation

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type PutRoute53ZoneRecordRequest struct {
	ZoneName   string        `json:"zoneName,omitempty"`
	RecordName string        `json:"recordName,omitempty"`
	RecordType string        `json:"recordType,omitempty"`
	TTL        time.Duration `json:"duration,omitempty"`
	Values     []string      `json:"values,omitempty"`
}

type PutRoute53ZoneRecordResponse struct {
	Action string `json:"action,omitempty"`
}

func (as *AutomationSet[Dependency]) PutRoute53ZoneRecords(ctx context.Context, req *PutRoute53ZoneRecordRequest) (*PutRoute53ZoneRecordResponse, error) {

	route53Cli := as.Dependency().GetRoute53Client()

	listZones, err := route53Cli.ListHostedZonesByName(ctx, &route53.ListHostedZonesByNameInput{
		DNSName:  aws.String(req.ZoneName),
		MaxItems: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}

	if len(listZones.HostedZones) == 0 {

	}

	hostedZoneId := listZones.HostedZones[0].Id
	recordType := types.RRType(strings.ToUpper(req.RecordType))

	listRecordPages := route53.NewListResourceRecordSetsPaginator(route53Cli, &route53.ListResourceRecordSetsInput{
		HostedZoneId:    hostedZoneId,
		StartRecordName: aws.String(req.RecordName),
		StartRecordType: recordType,
	})

	var recordSets []*types.ResourceRecordSet

	for listRecordPages.HasMorePages() {

		listRecordPage, err := listRecordPages.NextPage(ctx)
		if err != nil {

		}

		for _, recordSet := range listRecordPage.ResourceRecordSets {

			if req.RecordName == aws.ToString(recordSet.Name) {
				recordSets = append(recordSets, &recordSet)
			}

		}

	}

	res := &PutRoute53ZoneRecordResponse{
		Action: "noop",
	}

	changeRecordSetReq := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: hostedZoneId,
		ChangeBatch: &types.ChangeBatch{
			Comment: aws.String("Managed by Graviflow"),
			Changes: []types.Change{},
		},
	}

	resourceRecords := []types.ResourceRecord{}

	for _, val := range req.Values {
		resourceRecords = append(resourceRecords, types.ResourceRecord{
			Value: aws.String(val),
		})
	}

	found := false

	for _, recordSet := range recordSets {

		if recordSet.Type != recordType {
			continue
		}

		found = true
		skip := true

		if aws.ToInt64(recordSet.TTL) != int64(req.TTL.Seconds()) {
			recordSet.TTL = aws.Int64(int64(req.TTL.Seconds()))
			skip = false
		}

		if len(recordSet.ResourceRecords) != len(req.Values) {

			recordSet.ResourceRecords = resourceRecords

			skip = false

		} else {

			rrMap := make(map[string]interface{})

			for _, rr := range recordSet.ResourceRecords {
				rrMap[aws.ToString(rr.Value)] = nil
			}

			for _, val := range req.Values {

				if _, ok := rrMap[val]; !ok {
					skip = false
					break
				}

			}

			if !skip {
				recordSet.ResourceRecords = resourceRecords
			}

		}

		if skip {
			continue
		}

		res.Action = "updated"

		changeRecordSetReq.ChangeBatch.Changes = append(changeRecordSetReq.ChangeBatch.Changes, types.Change{
			Action:            types.ChangeActionUpsert,
			ResourceRecordSet: recordSet,
		})
	}

	if !found && len(recordSets) == 0 {
		res.Action = "created"

		changeRecordSetReq.ChangeBatch.Changes = []types.Change{{
			Action: types.ChangeActionCreate,
			ResourceRecordSet: &types.ResourceRecordSet{
				Name:            aws.String(req.RecordName),
				Type:            recordType,
				TTL:             aws.Int64(int64(req.TTL.Seconds())),
				ResourceRecords: resourceRecords,
			},
		}}

	}

	if len(changeRecordSetReq.ChangeBatch.Changes) > 0 {

		_, err := route53Cli.ChangeResourceRecordSets(ctx, changeRecordSetReq)
		if err != nil {
			return nil, err
		}

	}

	return res, nil

}
