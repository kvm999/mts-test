package shared

func Map[E any, T any](s []E, f func(E) T) []T {
	result := make([]T, len(s))

	for i := range s {
		result[i] = f(s[i])
	}

	return result
}

func Filter[E any](s []E, f func(E) bool) []E {
	var result []E

	for _, elem := range s {
		if f(elem) {
			result = append(result, elem)
		}
	}

	return result
}
