// 第一课只用标准库：先认识请求、状态码和耗时。
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:18090", "HTTP 监听地址")
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello, Go")
	})
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(250 * time.Millisecond)
		fmt.Fprintln(w, "hello, Go")
	})
	mux.HandleFunc("GET /error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "simulated failure", http.StatusInternalServerError)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	server := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("第一课 HTTP 服务：http://%s", *addr)
	log.Fatal(server.ListenAndServe())
}
