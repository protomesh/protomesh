package gateway

import (
	"sync"

	"github.com/protomesh/go-app/structures"
	typesv1 "github.com/protomesh/protomesh/proto/api/types/v1"
)

type policyMatcher struct {
	policy *typesv1.GatewayPolicy
	exact  bool
	tree   structures.RadixTree[*policyMatcher]
}

func newPolicyMatcher(exact bool) *policyMatcher {
	return &policyMatcher{
		exact: exact,
		tree:  structures.NewRadixTree[*policyMatcher](),
	}
}

func (m *policyMatcher) insert(policy *typesv1.GatewayPolicy, source any) {

	switch source := source.(type) {

	case *typesv1.GrpcSource:

		m.tree.Insert(source.MethodName, &policyMatcher{
			policy: policy,
		})

	case *typesv1.HttpSource:

		nested := m.tree.Match(source.Path)

		if nested == nil {
			nested := newPolicyMatcher(true)
			m.tree.Insert(source.Path, nested)
		}

		nested.insert(policy, source.Method)

	case typesv1.HttpMethod:

		switch source {

		case typesv1.HttpMethod_HTTP_METHOD_UNDEFINED:
			m.policy = policy

		default:
			m.tree.Insert(typesv1.HttpMethod_name[int32(source)], &policyMatcher{
				policy: policy,
			})

		}

	}

}

func (m *policyMatcher) drop(policy *typesv1.GatewayPolicy, source any) {

	switch source := source.(type) {

	case *typesv1.GrpcSource:

		nested := m.tree.Match(source.MethodName)

		if nested != nil {
			m.tree.Delete(source.MethodName)
		}

	case *typesv1.HttpSource:

		nested := m.tree.Match(source.Path)

		if nested != nil {
			nested.drop(policy, source.Method)
		}

	case typesv1.HttpMethod:

		switch source {

		case typesv1.HttpMethod_HTTP_METHOD_UNDEFINED:
			m.policy = nil

		default:
			m.tree.Delete(typesv1.HttpMethod_name[int32(source)])

		}

	}

}

func (m *policyMatcher) match(keys ...string) *typesv1.GatewayPolicy {

	key := keys[0]

	var nested *policyMatcher

	if m.exact {
		nested = m.tree.Match(key)
	} else {
		nested = m.tree.MatchLongest(key)
	}

	if nested == nil {
		return nil
	}

	if len(keys) == 1 {
		return nested.policy
	}

	if child := nested.match(keys[1:]...); child != nil {
		return child
	}

	return nil

}

type sourceMatcher struct {
	rwLock *sync.RWMutex

	exact  *policyMatcher
	prefix *policyMatcher
}

func newSourceMatcher() *sourceMatcher {
	return &sourceMatcher{
		rwLock: new(sync.RWMutex),
		exact:  newPolicyMatcher(true),
		prefix: newPolicyMatcher(false),
	}
}

func (m *sourceMatcher) addPolicy(policy *typesv1.GatewayPolicy) {

	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	source, exact := getPolicySource(policy)

	if exact {
		m.exact.insert(policy, source)
		return
	}

	m.prefix.insert(policy, source)

}

func (m *sourceMatcher) matchPolicy(keys ...string) *typesv1.GatewayPolicy {

	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	if policy := m.exact.match(keys...); policy != nil {
		return policy
	}

	if policy := m.prefix.match(keys...); policy != nil {
		return policy
	}

	return nil

}

func (m *sourceMatcher) dropPolicy(policy *typesv1.GatewayPolicy) {

	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	source, exact := getPolicySource(policy)

	if exact {
		m.exact.drop(policy, source)
		return
	}

	m.prefix.drop(policy, source)

}
