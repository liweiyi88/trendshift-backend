package sliceutils

func Chunk[T any](slice []T, size int) [][]T {
	var chunks [][]T

	for {
		if len(slice) == 0 {
			break
		}

		if len(slice) < size {
			size = len(slice)
		}

		chunks = append(chunks, slice[0:size])
		slice = slice[size:]
	}

	return chunks
}
