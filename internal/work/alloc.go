package work

// AllocHeavy создаёт n буферов по 4 КиБ, чтобы получить заметный alloc_space.
func AllocHeavy(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		buf := make([]byte, 4096)
		for j := range buf {
			buf[j] = byte(i + j)
		}
		sum += int(buf[i%len(buf)])
	}
	return sum
}