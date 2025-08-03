package main

import "net/http"
import "github.com/gorilla/mux"

func main() {
	router := mux.NewRouter()
	_ = router.HandleFunc("/hello", hello-handler)
	_ = http.ListenAndServe(":8080", router)
}
