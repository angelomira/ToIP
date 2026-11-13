package work

import "testing"

// Бенчмарк «до оптимизации»: n=30, чтобы бенчмарк шёл разумное время.
func BenchmarkFib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Fib(30)
	}
}

// Бенчмарк «после оптимизации».
// n объявлен как переменная (а не константа), чтобы компилятор
// не свернул FibFast(n) в константу на этапе компиляции.
func BenchmarkFibFast(b *testing.B) {
	n := 30
	for i := 0; i < b.N; i++ {
		_ = FibFast(n)
	}
}