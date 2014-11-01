package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

// Handler for admin api
type AdminHandler struct {
	Target *BlueGreenHandler
}

func (ah *AdminHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Admin] serving %s", req.URL)

	fmt.Fprintf(w, "I am the admin api! ENV is %s\n", ah.Target.Environment)

	// toggle environments
	if ah.Target.Environment == BLUE {
		ah.Target.Environment = GREEN
	} else {
		ah.Target.Environment = BLUE
	}

	fmt.Fprintf(w, "Now I switched to %s\n", ah.Target.Environment)
	log.Printf("[Admin] switched to: %s\n", ah.Target.Environment)
}

// Handler for blue green switcher
type BlueGreenHandler struct {
	Blue        *httputil.ReverseProxy
	Green       *httputil.ReverseProxy
	Environment string
}

func makeBlueGreen(blueString string, greenString string) *BlueGreenHandler {
	bgh := BlueGreenHandler{Environment: BLUE}

	blueUrl, _ := url.Parse(blueString)
	greenUrl, _ := url.Parse(greenString)

	bgh.Blue = httputil.NewSingleHostReverseProxy(blueUrl)
	bgh.Green = httputil.NewSingleHostReverseProxy(greenUrl)

	return &bgh
}

func (bgh *BlueGreenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Main]  serving from %s: %s", bgh.Environment, req.URL.Path)

	if bgh.Environment == BLUE {
		bgh.Blue.ServeHTTP(w, req)
	} else {
		bgh.Green.ServeHTTP(w, req)
	}
}

// Handler for alternate requests (s-out environment)
type InverseHandler struct {
	Target *BlueGreenHandler
}

func (ih *InverseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Alt]   serving from NOT %s: %s", ih.Target.Environment, req.URL.Path)

	if ih.Target.Environment == BLUE {
		ih.Target.Green.ServeHTTP(w, req)
	} else {
		ih.Target.Blue.ServeHTTP(w, req)
	}
}

func main() {
	// command line flags
	blueBackend := flag.String("blue", "http://localhost", "url to blue environment")
	greenBackend := flag.String("green", "http://localhost", "url to green environment")

	blueGreenPort := flag.String("main", ":8080", "listen address for main http handler")
	inversePort := flag.String("alt", ":8181", "listen address for alternate/test http handler")
	adminPort := flag.String("admin", ":5000", "listen address for admin api")

	flag.Parse()

	bgHandler := makeBlueGreen(*blueBackend, *greenBackend)
	testHandler := InverseHandler{Target: bgHandler}
	adminHandler := AdminHandler{Target: bgHandler}

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

		log.Fatal(http.ListenAndServe(*adminPort, &adminHandler))
		wg.Done()
	}()

	wg.Wait()
}
