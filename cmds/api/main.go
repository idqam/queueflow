package api

import "net/http"




func main() {
	server := &http.Server{
		Addr:    ":8080",
		Handler: NewRouter(),
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
