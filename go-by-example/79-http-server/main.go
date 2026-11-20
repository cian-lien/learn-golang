package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "书籍编号 ", r.PathValue("id"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/books/42")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		panic(err)
	}
	fmt.Println()
}
