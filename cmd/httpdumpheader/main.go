package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	d, err := json.Marshal(r.Header)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(err.Error()))

		log.Println(err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(d)

	log.Println(string(d) + "\n")
}

func main() {
	var listen string

	flag.StringVar(&listen, "l", ":8080", "listen address:port")
	flag.Parse()

	http.HandleFunc("/", indexHandler)

	log.Println("server listen on:", listen)

	_ = http.ListenAndServe(listen, nil) // nolint: gosec
}
