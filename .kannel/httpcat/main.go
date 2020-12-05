package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"unicode/utf8"

	"github.com/mdouchement/smpp/smpp/pdu/pdutext"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Println("ERROR: could not dump request")
		}
		fmt.Println(string(dump))

		query := r.URL.Query()
		if len(query) > 0 {
			fmt.Println("=== Parsed query ===")

			for k := range query {
				v := []byte(query.Get(k))
				if !utf8.Valid(v) {
					v = pdutext.UCS2(v).Decode()
				}

				fmt.Fprintf(os.Stdout, "%s\t\t%s\n", k, v)
			}

			fmt.Println("====================")
		}

		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Listening HTTP on :8888")
	err := http.ListenAndServe(":8888", nil)
	if err != nil {
		panic(err)
	}
}
