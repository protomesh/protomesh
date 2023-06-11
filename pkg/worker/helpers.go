package worker

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/protomesh/protomesh/pkg/resource"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

func workflowIdFromTrigger(trigger *typesv1.Trigger) (string, error) {

	switch idSuffix := trigger.IdSuffix.(type) {

	case *typesv1.Trigger_ExactIdSuffix:
		return strings.Join([]string{trigger.IdPrefix, idSuffix.ExactIdSuffix}, ""), nil

	case *typesv1.Trigger_IdSuffixBuilder:

		switch idSuffix.IdSuffixBuilder {

		case typesv1.Trigger_ID_BUILDER_ONLY_PREFIX:
			return trigger.IdPrefix, nil

		case typesv1.Trigger_ID_BUILDER_RANDOM:
			randomId, err := uuid.NewRandom()
			if err != nil {
				return "", err
			}

			return strings.Join([]string{trigger.IdPrefix, randomId.String()}, ""), nil

		case typesv1.Trigger_ID_BUILDER_UNIQUE:
			uniqueId := uuid.NewSHA1(resource.WorkflowIdNamespace, []byte(trigger.IdPrefix))
			return strings.Join([]string{trigger.IdPrefix, uniqueId.String()}, ""), nil

		}

	}

	return "", errors.New("Invalid ID suffix")

}
