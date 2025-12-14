package utils

func Default[T comparable](val, def T) T {
	var zero T
	if val == zero {
		return def
	}
	return val
}
