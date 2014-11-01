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

type AdminHandler struct {
	Target *BlueGreenHandler
}

func (ah *AdminHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Admin] serving %s env: %s\n", req.URL, ah.Target.Environment)

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
	log.Println("[BlueGreen] serving ", req.URL, " from: ", bgh.Environment)

	if bgh.Environment == BLUE {
		bgh.Blue.ServeHTTP(w, req)
		//w.Write([]byte("I am Blue!!"))
	} else {
		bgh.Green.ServeHTTP(w, req)
		//w.Write([]byte("I am Green!!"))
	}
}

type InverseHandler struct {
	Target *BlueGreenHandler
}

func (ih *InverseHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("[Inverse] serving ", req.URL, " from: NOT ", ih.Target.Environment)

	if ih.Target.Environment == BLUE {
		ih.Target.Green.ServeHTTP(w, req)
	} else {
		ih.Target.Blue.ServeHTTP(w, req)
	}
}

func main() {
	// command line flags
	blueBackend := flag.String("blue", "", "url to blue environment")
	greenBackend := flag.String("green", "", "url to green environment")
	flag.Parse()

	bgHandler := makeBlueGreen(*blueBackend, *greenBackend)
	testHandler := InverseHandler{Target: bgHandler}
	adminHandler := AdminHandler{Target: bgHandler}

	wg := &sync.WaitGroup{}

	// Start the main handler
	wg.Add(1)
	go func() {
		blueGreenPort := ":5051"
		log.Println("[BlueGreen] Started listening on ", blueGreenPort)

		log.Fatal(http.ListenAndServe(blueGreenPort, bgHandler))
		wg.Done()
	}()

	// Start the inverse handler
	wg.Add(1)
	go func() {
		inversePort := ":5150"
		log.Println("[Inverse] Started listening on ", inversePort)

		log.Fatal(http.ListenAndServe(inversePort, &testHandler))
		wg.Done()
	}()

	// Start the admin handler
	wg.Add(1)
	go func() {
		adminPort := ":5555"
		log.Println("[Admin] Started listening on ", adminPort)

		log.Fatal(http.ListenAndServe(adminPort, &adminHandler))
		wg.Done()
	}()

	wg.Wait()
}
