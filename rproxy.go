package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

const (
	BLUE  = "blue"
	GREEN = "green"
)

// response models
type AdminResponse struct {
	Command string `json:"command"`
	Status  string `json:"status"`
	Result  string `json:"result,omitempty"`
	Blue    string `json:"blue,omitempty"`
	Green   string `json:"green,omitempty"`
}

// Handler for admin api
type AdminHandler struct {
	Target *BlueGreenHandler
}

func (ah *AdminHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Admin] serving %s", req.URL)

	paths := strings.Split(req.URL.String(), "/")

	command := "status"
	param := ""

	if len(paths) > 1 {
		command = paths[1]
		if command == "" {
			command = "status"
		}
	}
	if len(paths) > 2 {
		param = paths[2]
	}

	log.Println("  [Admin] command: " + command)
	log.Println("  [Admin]   param: " + param)

	res := AdminResponse{Command: command}

	initial := ah.Target.Environment

	if command == "toggle" {

		msg := "switched from " + initial

		// toggle environments
		if ah.Target.Environment == BLUE {
			ah.Target.Environment = GREEN
		} else {
			ah.Target.Environment = BLUE
		}

		res.Result = msg

		log.Printf("  [Admin] switched from: %s to: %s\n", initial, ah.Target.Environment)
	} else if command == "switch" {

		msg := "switched from " + initial

		// toggle environments
		log.Printf("  [Admin] trying to switch to: %s\n", param)
		if param == BLUE {
			ah.Target.Environment = BLUE
		} else if param == GREEN {
			ah.Target.Environment = GREEN
		} else {
			msg = "invalid environment"
		}

		res.Result = msg
	}

	res.Status = ah.Target.Environment

	output, _ := json.MarshalIndent(res, " ", "  ")

	fmt.Fprintf(w, string(output))

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
