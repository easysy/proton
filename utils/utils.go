package utils

func ClonePointer[T any](src *T) *T {
	if src == nil {
		return nil
	}
	dst := new(T)
	*dst = *src
	return dst
}

func ContainsDuplicates[S ~[]E, E comparable](s S) bool {
	encountered := make(map[E]bool, len(s))

	for i := range s {
		if encountered[s[i]] {
			return true
		}
		encountered[s[i]] = true
	}

	return false
}
