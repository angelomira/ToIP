package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"example.com/pprof-lab/internal/work"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	// именованные профили (allocs, block, goroutine, heap, mutex, threadcreate)
	mux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

	// «Медленный» эндпоинт: рекурсивный Фибоначчи.
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		defer work.TimeIt("Fib(38)")()
		n := 38
		res := work.Fib(n)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%d\n", res)
	})

	// «Быстрый» эндпоинт: итеративный Фибоначчи (после оптимизации).
	mux.HandleFunc("/work/fast", func(w http.ResponseWriter, r *http.Request) {
		defer work.TimeIt("FibFast(38)")()
		n := 38
		res := work.FibFast(n)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%d\n", res)
	})

	// Эндпоинт с аллокациями — для Heap-профиля.
	mux.HandleFunc("/alloc", func(w http.ResponseWriter, r *http.Request) {
		res := work.AllocHeavy(1000)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "alloc ok, checksum=%d\n", res)
	})

	// Эндпоинт с мьютексом — для Block/Mutex-профилей (бонус).
	mux.HandleFunc("/locked", func(w http.ResponseWriter, r *http.Request) {
		res := work.LockedWork(35)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%d\n", res)
	})

	log.Println("Server on :8080; pprof on /debug/pprof/")
	log.Fatal(http.ListenAndServe(":8080", mux))
}