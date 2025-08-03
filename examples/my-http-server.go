package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteString("Hello from Vex HTTP Server!")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteString("{\"status\": \"ok\", \"message\": \"Vex server is running!\"}")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", helloHandler)
	router.HandleFunc("/status", statusHandler)
	
	fmt.Println("🚀 Vex HTTP Server starting on :8080...")
	fmt.Println("Try: curl http://localhost:8080/")
	fmt.Println("Try: curl http://localhost:8080/status")
	
	http.ListenAndServe(":8080", router)
}