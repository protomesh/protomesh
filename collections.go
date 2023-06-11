package protomesh

type TypeSet[K comparable] Set[K, K]

func (t TypeSet[K]) FromSlice(keys ...K) {
	for _, key := range keys {
		t[key] = key
	}
}

type Set[K comparable, V any] map[K]V

func (s Set[K, V]) Set(key K, value V) {
	s[key] = value
}

func (s Set[K, V]) Has(key K) bool {
	_, ok := s[key]
	return ok
}

func (s Set[K, V]) Del(key K) {
	delete(s, key)
}

func (s Set[K, V]) ToValueSlice() []V {

	slc := make([]V, 0)

	for _, value := range s {
		slc = append(slc, value)
	}

	return slc

}
