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

func MapGetValue[K comparable, V any](m map[K]any, k K) (v V, ok bool) {
	if m != nil {
		v, ok = m[k].(V)
	}
	return
}

func MapGetValueSilent[K comparable, T any](m map[K]any, k K) T {
	t, _ := MapGetValue[K, T](m, k)
	return t
}
