package generic

func ClonePointer[T any](src *T) *T {
	if src == nil {
		return nil
	}
	dst := new(T)
	*dst = *src
	return dst
}

func TakePointer[T any](v T) *T {
	return &v
}

func Encounter[S ~[]E, E comparable](s S) func(E) bool {
	encountered := make(map[E]bool, len(s))
	return func(e E) bool {
		if encountered[e] {
			return true
		}
		encountered[e] = true
		return false
	}
}

func GetFromMap[K comparable, T any](m map[K]any, k K) (T, bool) {
	var t T

	if m == nil {
		return t, false
	}

	v, ok := m[k]
	if ok && v != nil {
		t, ok = v.(T)
	}

	return t, ok
}
