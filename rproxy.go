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
	log.Println("[Admin] serving ", req.URL, " env: ", ah.Target.Environment)

	fmt.Fprintf(w, "I am the admin api! ENV is %s\n", ah.Target.Environment)

	// toggle environments
	if ah.Target.Environment == BLUE {
		ah.Target.Environment = GREEN
	} else {
		ah.Target.Environment = BLUE
	}

	fmt.Fprintf(w, "Now I switched to %s\n", ah.Target.Environment)
	log.Println("[Admin] switched to: %s\n", ah.Target.Environment)
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

func (bgh *BlueGreenHandler) makeInverseHandler() *BlueGreenHandler {
	return &BlueGreenHandler{
		Environment: GREEN, // TODO need to share this flag too
		Blue:        bgh.Green,
		Green:       bgh.Blue,
	}
}

func (bgh *BlueGreenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("[Blue Green] serving ", req.URL, " from: ", bgh.Environment)

	if bgh.Environment == BLUE {
		bgh.Blue.ServeHTTP(w, req)
		//w.Write([]byte("I am Blue!!"))
	} else {
		bgh.Green.ServeHTTP(w, req)
		//w.Write([]byte("I am Green!!"))
	}
}

func main() {
	// command line flags
	blueBackend := flag.String("blue", "", "url to blue environment")
	greenBackend := flag.String("green", "", "url to green environment")
	flag.Parse()

	bgHandler := makeBlueGreen(*blueBackend, *greenBackend)
	//testHandler := bgHandler.makeInverseHandler()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		adminPort := ":5000"
		log.Println("[Admin] Started listening on ", adminPort)

		log.Fatal(http.ListenAndServe(adminPort, &AdminHandler{bgHandler}))
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		blueGreenPort := ":5050"
		log.Println("[Blue Green] Started listening on ", blueGreenPort)

		log.Fatal(http.ListenAndServe(blueGreenPort, bgHandler))
		wg.Done()
	}()

	wg.Wait()
}
