package utils

func ClonePointer[T any](src *T) *T {
	if src == nil {
		return nil
	}
	dst := new(T)
	*dst = *src
	return dst
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
