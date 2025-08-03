package main

import "net/http"
import "github.com/gorilla/mux"
import "fmt"

func main() {
	// Registered macro: defn
	hello-handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteString("Hello from Ultimate AI-Friendly Vex!")
	}
	status-handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteString("{\"status\": \"ok\", \"language\": \"vex\", \"syntax\": \"slash-notation\"}")
	}
	router := mux.NewRouter()
	_ = router.HandleFunc("/", hello-handler)
	_ = router.HandleFunc("/status", status-handler)
	fmt.Println("🚀 Ultimate AI-Friendly Vex HTTP Server")
	fmt.Println("📡 Serving on http://localhost:8080")
	fmt.Println("✨ Using pure slash notation syntax")
	http.ListenAndServe(":8080", router)
}
