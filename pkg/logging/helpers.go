package logging

import (
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

// func idFromServiceMeshNode(namespace string, node *typesv1.ServiceMesh_Node) string {

// 	hash := sha256.New()

// 	switch node := node.Node.(type) {

// 	case *typesv1.ServiceMesh_Node_HttpIngress:
// 		hash.Write([]byte(node.HttpIngress.XdsClusterName))
// 		hash.Write([]byte(node.HttpIngress.IngressName))

// 	case *typesv1.ServiceMesh_Node_RoutingPolicy:
// 		hash.Write([]byte(node.RoutingPolicy.XdsClusterName))
// 		hash.Write([]byte(node.RoutingPolicy.Domain))
// 		hash.Write([]byte(node.RoutingPolicy.IngressName))

// 	case *typesv1.ServiceMesh_Node_Service:
// 		hash.Write([]byte(node.Service.ServiceName))

// 	default:

// 	}

// 	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

// }

func LogResource(res *typesv1.Resource, kvs ...interface{}) []interface{} {
	return append([]interface{}{
		"resourceNamespace", res.Namespace,
		"resourceId", res.Id,
		"resourceName", res.Name,
	}, kvs...)
}

func LogTrigger(workTrigger *typesv1.Trigger, kvs ...interface{}) []interface{} {
	return append([]interface{}{
		"workflowName", workTrigger.Name,
		"taskQueue", workTrigger.TaskQueue,
		"workflowIdPrefix", workTrigger.IdPrefix,
	}, kvs...)
}
