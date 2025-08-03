package main

import "net/http"
import "github.com/gorilla/mux"
import "fmt"

func main() {
	// Registered macro: defn
	hello-handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteString("Hello from Vex HTTP Server using defn macro!")
	}
	status-handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteString("{\"status\": \"ok\", \"server\": \"vex\"}")
	}
	router := mux.NewRouter()
	_ = router.HandleFunc("/", hello-handler)
	_ = router.HandleFunc("/status", status-handler)
	_ = fmt.Println
	_ = "🚀 Vex HTTP Server with defn macro starting on :8080..."
	_ = http.ListenAndServe(":8080", router)
}
