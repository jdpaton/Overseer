package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/jmhodges/levigo"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
)

const (
	port = "5600"
)

func runCmd(program string, args string) int {

	log.Printf("New request to run: `%s %s`", program, args)
	cmd := exec.Command(program, args)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()

	if err != nil {
		log.Printf("Failed to run command: %v", err)
		log.Printf("Command stdout: %s\n", stdout.String())
		log.Printf("Command stderr: %s\n", stderr.String())
		return -1
	}

	return cmd.Process.Pid

}

func reqRunCmd(program string, args string) int {
	pid := runCmd(program, args)
	return pid
}

func clientReqCmd(program string, args string) {
	values := make(url.Values)

	values.Set("program", program)
	values.Set("args", args)

	r, err := http.PostForm(fmt.Sprintf("http://127.0.0.1:%s/new", port), values)
	if err != nil {
		log.Fatal(fmt.Sprintf("error requesting new program: %s\n", program), err)
	}

	defer r.Body.Close()

	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("%s", body)

}

func handleReq(w http.ResponseWriter, r *http.Request, db *levigo.DB) {
	/* Start */
	if r.Method == "POST" && r.URL.Path[1:] == "new" {
		pid := reqRunCmd(r.FormValue("program"), r.FormValue("args"))
		fmt.Fprintf(w, "PID: %d", pid)
	}
	/* Stop */
	if r.Method == "POST" && r.URL.Path[1:] == "stop" {
		pid, err := strconv.Atoi(r.FormValue("pid"))

		err = removeProc(pid, db)
		if err != nil {
			fmt.Fprintf(w, "Error stopping pid %d, %v\n", pid, err)
			return
		}
		fmt.Fprintf(w, "Stopped PID %d", pid)
	}
}

func startServer(db *levigo.DB) {
	log.Printf("Starting server on port %s", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleReq(w, r, db)
	})
	http.ListenAndServe("127.0.0.1:"+port, nil)
}

func main() {
	var program = flag.String("program", "", "the full path to the command to be run")
	var args = flag.String("args", "", "a string of all arguments to pass to the program")
	var startserver = flag.Bool("server", false, "start the overseer server")

	flag.Parse()

	if *startserver == true {
		db, dberr := InitDB()

		if dberr != nil {
			log.Fatal("Cannot open database", dberr)
		}
		defer db.Close()

		startServer(db)
	} else if *program == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else if *program != "" && *args != "" {
		clientReqCmd(*program, *args)
	}
}
