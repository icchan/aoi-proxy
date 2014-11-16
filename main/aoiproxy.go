package main

import (
	"flag"
	"github.com/icchan/aoi-proxy"
	"log"
	"net/http"
	"sync"
)

func main() {
	// command line flags
	blueBackend := flag.String("blue", "http://localhost:7001", "url to blue environment")
	greenBackend := flag.String("green", "http://localhost:7002", "url to green environment")

	blueGreenPort := flag.String("main", ":8080", "listen address for main http handler")
	inversePort := flag.String("alt", ":8081", "listen address for alternate/test http handler")
	adminPort := flag.String("admin", ":5000", "listen address for admin api")

	flag.Parse()

	bgHandler := aoiproxy.NewBlueGreen(*blueBackend, *greenBackend)
	testHandler := aoiproxy.InverseHandler{Target: bgHandler}
	adminHandler := aoiproxy.NewAdminHandler(bgHandler)

	wg := &sync.WaitGroup{}

	// Start the main handler
	wg.Add(1)
	go func() {
		log.Println("[Main]  Started listening on", *blueGreenPort)

		log.Fatal(http.ListenAndServe(*blueGreenPort, bgHandler))
		wg.Done()
	}()

	// Start the inverse handler
	wg.Add(1)
	go func() {
		log.Println("[Alt]   Started listening on", *inversePort)

		log.Fatal(http.ListenAndServe(*inversePort, &testHandler))
		wg.Done()
	}()

	// Start the admin handler
	wg.Add(1)
	go func() {
		log.Println("[Admin] Started listening on", *adminPort)

		log.Fatal(http.ListenAndServe(*adminPort, adminHandler))
		wg.Done()
	}()

	wg.Wait()
}
