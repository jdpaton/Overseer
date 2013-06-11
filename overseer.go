package main

import (
	"flag"
	"fmt"
	"github.com/jmhodges/levigo"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
)

const (
	default_port = "5600"
)

func runCmd(id, program, args string) int {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	log_dir := path.Join(usr.HomeDir, ".overseer", "logs")
	stdout_file, err := os.Create(path.Join(log_dir, id+"-stdout.log"))
	stderr_file, err := os.Create(path.Join(log_dir, id+"-stderr.log"))

	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return -1
	}

	log.Printf("New request to run: `%s %s`", program, args)
	cmd := exec.Command(program, args)

	cmd.Stdout = stdout_file
	cmd.Stderr = stderr_file

	err = cmd.Start()

	if err != nil {
		log.Printf("Failed to run command: %v", err)
		return -1
	}

	return cmd.Process.Pid

}

func reqRunCmd(program, args string) (string, int) {
	id := randString(8)
	pid := runCmd(id, program, args)
	return id, pid
}

func clientReqCmd(program, args, port string) {
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
		id, pid := reqRunCmd(r.FormValue("program"), r.FormValue("args"))
		fmt.Fprintf(w, "ID: %s PID: %d", id, pid)
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

func startServer(db *levigo.DB, port string) {
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
	var port = flag.String("port", default_port, "server listen port")

	flag.Parse()

	if *startserver == true {
		db, dberr := InitDB()

		if dberr != nil {
			log.Fatal("Cannot open database", dberr)
		}
		defer db.Close()

		startServer(db, *port)
	} else if *program == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else if *program != "" {
		clientReqCmd(*program, *args, *port)
	}
}
