package main

import (
	//"encoding/json"
	"fmt"
	"flag"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var (
	path *string
	token *string
)

func rootRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.FormValue("token") != *token {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	script := r.URL.Path[1:]
	if strings.Contains(script, "/") {
		msg := fmt.Sprintf("'%s' contains slashes, and is not allowed.", script)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	fullpath := fmt.Sprintf("%s/%s", *path, script)

	cmd := exec.Command(fullpath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Fprintf(w, "%s%v\n", output, err)
		return
	}

	fmt.Fprintf(w, "%s", output)
}

func main() {
	path = flag.String("path", "", "path to scripts directory")
	token = flag.String("token", "", "token")

	flag.Parse()

	if (*path == "") {
		fmt.Printf("dead\n")
		return
	}

	http.HandleFunc("/", rootRoute)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
