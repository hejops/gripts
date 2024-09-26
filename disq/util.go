package main

func Must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}

func Keys[T comparable, U any](m map[T]U) []T {
	var keys []T
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
