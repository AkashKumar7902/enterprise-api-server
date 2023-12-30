package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"

	// "fmt"
	"log"
	"net/http"

	// "reflect"
	"time"

	"github.com/gorilla/mux"
)

var commands = []string{
	"keploy record -c /test-app-url-shortener",
}

var jid = []string{}

var flag bool = false

func returnCmd(w http.ResponseWriter, r *http.Request) {
	fmt.Println(flag)
	if !flag {
		json.NewEncoder(w).Encode(commands)
		flag = true
		return
	}
	ticker := time.Tick(2000 * time.Second)
	<-ticker
	json.NewEncoder(w).Encode(commands)
}

func storeJID(w http.ResponseWriter, r *http.Request) {
	var jidBody string
	_ = json.NewDecoder(r.Body).Decode(&jidBody)
	jid = append(jid, jidBody)
	fmt.Println(jid)
}

type cmdErr struct {
	JID string `json:"jid"`
	Msg string `json:"msg"`
}

func stop(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Stopping...")
	var body cmdErr
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Println("Error decoding body", err)
		return
	}
	fmt.Println(body)

	w.WriteHeader(http.StatusOK)
}

func handleAgentStop(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Agent stopped")
	w.WriteHeader(http.StatusOK)
}

func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("keploy_cmd_logfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
}

func handleRequests() {
	r := mux.NewRouter()

	r.HandleFunc("/cmd", returnCmd).Methods("GET")
	r.HandleFunc("/jid", storeJID).Methods("POST")
	r.HandleFunc("/cmdError", stop).Methods("POST")
	r.HandleFunc("/log", upload).Methods("POST")
	r.HandleFunc("/agentstopped", handleAgentStop).Methods("GET")


	log.Fatal(http.ListenAndServe(":8000", r))
}

func main() {

	go func() {
		ticker := time.Tick(5 * time.Second)
		<-ticker
		if len(jid) > 0 {
			fmt.Print("i am here\n")
			commands = append(commands, "StopCmd --jid " + jid[0])
		}
	}()

	previousArray := make([]string, len(jid))
	copy(previousArray, jid)

	go func() {
		if !reflect.DeepEqual(jid, previousArray) {
			fmt.Println("Array has changed:", jid)
			copy(previousArray, jid) // Update the copy with the current array
		} else {
			fmt.Println("No changes detected")
		}
	}()

	handleRequests()
}