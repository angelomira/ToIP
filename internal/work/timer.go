package work

import (
	"log"
	"time"
)

// TimeIt возвращает функцию, которую нужно вызвать отложенно (defer).
// Пример: defer work.TimeIt("Fib(38)")()
func TimeIt(name string) func() {
	start := time.Now()
	return func() {
		log.Printf("%s took %s", name, time.Since(start))
	}
}