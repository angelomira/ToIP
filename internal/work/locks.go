package work

import (
	"runtime"
	"sync"
)

var (
	mu    sync.Mutex
	store = make(map[int]int)
)

// init включается один раз при старте приложения.
func init() {
	runtime.SetBlockProfileRate(1)       // Block profile: все блокировки >= 1 нс
	runtime.SetMutexProfileFraction(1)   // Mutex profile: все contended-мьютексы
}

// LockedWork — работа под мьютексом с «медленным» кэшем:
// специально держим блокировку во время тяжёлого вычисления,
// чтобы получить ненулевой mutex/block профиль.
func LockedWork(n int) int {
	mu.Lock()
	defer mu.Unlock()

	if v, ok := store[n]; ok {
		return v
	}
	v := Fib(n) // держим мьютекс во время вычисления — плохая практика, нужна для демонстрации
	store[n] = v
	return v
}