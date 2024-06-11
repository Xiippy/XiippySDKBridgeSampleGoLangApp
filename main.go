package main

import (
	"XiippySDKBridgeSampleGoLangApp/handlers"
	"net/http"
)

func main() {
	http.HandleFunc("/", handlers.HomeHandler)
	fs := http.FileServer(http.Dir("./wwwroot"))
	http.Handle("/wwwroot/", http.StripPrefix("/wwwroot/", fs))
	http.ListenAndServe(":3000", nil)
}
