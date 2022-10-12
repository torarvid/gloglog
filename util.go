package main

func contains[E comparable](haystack []E, needle E) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
