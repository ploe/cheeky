package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"flag"
	"log"
	"log/syslog"
	"net/http"
	"os/exec"
	"strings"
)

var (
	path *string
	port *int
	tag *string
	tls *string
	token *string
)

type SlackResponse struct {
	Type string `json:"response_type"`
	Text string `json:"text"`
}

func validRequest(w http.ResponseWriter, r *http.Request) (string, int) {
	script := r.FormValue("text")

	if r.FormValue("token") != *token {
		return http.StatusText(http.StatusForbidden), http.StatusForbidden
	}

	if r.Method != "POST" {
		return http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed
	}

	if script == "" {
		text := fmt.Sprintf("no script specified")
		return text, http.StatusNotFound
	}

	if strings.ContainsAny(script, "/;&") {
		text := fmt.Sprintf("'%s' cannot contain characters: / ; &", script)
		return text, http.StatusNotFound
	}


	return "", http.StatusOK
}

func execCommand(user, script, url string) {
	fullpath := fmt.Sprintf("%s/%s", *path, script)

	cmd := exec.Command(fullpath)
	output, err := cmd.CombinedOutput()

	/* if the command failed we should include the status of the script */
	slack := SlackResponse{Type:"in-channel"}
	if err != nil {
		slack.Text = fmt.Sprintf("%s%v", output, err)
	} else {
		slack.Text = string(output)
	}

	reply, _ := json.Marshal(slack)

	http.Post(url, "application/json", bytes.NewReader(reply))

	log.Printf("%s %s %s %s",
		script,
		user,
		url,
		reply,
	)

}

func rootRoute(w http.ResponseWriter, r *http.Request) {
	script := r.FormValue("text")
	user := r.FormValue("user_name")

	text, status := validRequest(w, r)

	if status != http.StatusOK {
		log.Printf("%d:%s:%s \"%s\"",
			status,
			script,
			user,
			text,
		)
		http.Error(w, text, status)
		return
	}

	go execCommand(user, script, r.FormValue("response_url"))

	fmt.Fprintf(w, "Hey %s, I'm running '%s' - gimme a mo...", user, script)
}

func main() {
	path = flag.String("path", "", "path to scripts directory")
	port = flag.Int("port", 8080, "port")
	tag = flag.String("tag", "", "tag for the syslog. if not set uses stdout instead")
	tls = flag.String("tls", "", "path to certs generated by certbot")
	token = flag.String("token", "", "secret token for the api")

	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	if *tag != "" {
		log.SetFlags(0)
		logger, _ := syslog.New(syslog.LOG_NOTICE, *tag)
		log.SetOutput(logger)
	}

	if (*path == "") {
		log.Fatal("the flag path is required")
		return
	}

	http.HandleFunc("/", rootRoute)
	if *tls != "" {
		fullchain := fmt.Sprintf("%s/fullchain.pem", *tls)
		privkey := fmt.Sprintf("%s/privkey.pem", *tls)
		log.Fatal(http.ListenAndServeTLS(addr, fullchain, privkey, nil))
	}

	log.Fatal(http.ListenAndServe(addr, nil))
}
