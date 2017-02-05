package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("log middleware begin")
		next.ServeHTTP(w, r)
		fmt.Println("log middleware end")
	})
}

func FilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("filter middleware begin")
		if r.URL.Path != "/" {
			return
		}

		next.ServeHTTP(w, r)
		fmt.Println("filter middleware end")
	})
}

func EnforceXmlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		b := new(bytes.Buffer)
		b.ReadFrom(r.Body)
		if http.DetectContentType(b.Bytes()) != "text/xml; charset=utf-8" {
			http.Error(w, http.StatusText(415), 415)
			return
		}

		next.ServeHTTP(w, r)
	})
}

//def final handle
func final(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

func main() {
	finalHandle := http.HandlerFunc(final)

	http.Handle("/", LogMiddleware(FilterMiddleware(EnforceXmlMiddleware(finalHandle))))
	http.ListenAndServe(":3000", nil)
}
