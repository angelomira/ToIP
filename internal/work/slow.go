package work

// Fib — неоптимальный рекурсивный Фибоначчи: демонстрация CPU-нагрузки.
// Сложность O(2^n) по числу вызовов.
func Fib(n int) int {
	if n < 2 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}

// FibFast — итеративная версия, O(n) по времени и O(1) по памяти.
func FibFast(n int) int {
	if n < 2 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}