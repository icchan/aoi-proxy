package aoiproxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// Handler for admin api
type AdminHandler struct {
	Target      *BlueGreenHandler
	FileHandler http.Handler
}

// response models
type AdminResponse struct {
	Command string `json:"command"`
	Status  string `json:"status"`
	Result  string `json:"result,omitempty"`
	Blue    string `json:"blue,omitempty"`
	Green   string `json:"green,omitempty"`
}

func NewAdminHandler(bg *BlueGreenHandler) *AdminHandler {
	return &AdminHandler{Target: bg, FileHandler: http.FileServer(http.Dir("www"))}
}

func (ah *AdminHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Admin] serving %s", req.URL)

	if req.URL.String() == "/" && ah.FileHandler != nil {
		// serve the webpage
		ah.FileHandler.ServeHTTP(w, req)
		return
	}

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

	res := AdminResponse{
		Command: command,
		Blue:    ah.Target.Blue.Target,
		Green:   ah.Target.Green.Target,
	}

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
