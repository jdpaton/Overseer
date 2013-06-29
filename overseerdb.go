package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jmhodges/levigo"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func InitDB() (*levigo.DB, error) {
	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(1024 * 1024))
	opts.SetCreateIfMissing(true)

	homedir := os.Getenv("HOME")
	dbdir := homedir + "/.overseer/db"
	os.MkdirAll(dbdir, 0700)

	db, err := levigo.Open(dbdir, opts)

	if err != nil {
		fmt.Println("Failed to open database", err)
		return nil, err
	}

	procs, err := ListProcs(db)
	for p, status := range procs {
		log.Printf("Proc ID: %d, status: %d", p, status)
		if isProcAlive(p) == false {
			setProcStatus(db, p, PROC_STOPPED)
		} else {
			setProcStatus(db, p, PROC_ALIVE)
		}
	}
	return db, nil
}

func setProcStatus(db *levigo.DB, pid, status int) {
	wo := levigo.NewWriteOptions()
	key := "status:" + strconv.Itoa(pid)
	db.Put(wo, []byte(key), []byte(strconv.Itoa(status)))
}

func isProcAlive(pid int) bool {
	// Doesn't work as expected on POSIX systems
	// - https://groups.google.com/forum/#!topic/golang-nuts/WtIsS9dzy68
	//_, err := os.FindProcess(pid)
	if err := syscall.Kill(pid, 0); err != nil {
		return false
	} else {
		return true
	}

}

func removeProc(pid int, db *levigo.DB) error {
	wo := levigo.NewWriteOptions()
	err := db.Delete(wo, []byte(strconv.Itoa(pid)))
	if err != nil {
		return err
	}
	return killProc(pid)

}

func killProc(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = proc.Kill()
	return err
}

func AddProc(procID int, db *levigo.DB) error {
	ro := levigo.NewReadOptions()
	wo := levigo.NewWriteOptions()

	data, err := db.Get(ro, []byte("procs"))
	spdata := bytes.Split(data, []byte(":"))

	for i, e := range spdata {

		if string(e) != "" {
			fmt.Println("ProcID: #", i, string(e))
			pid, err := strconv.Atoi(string(e))
			if err != nil {
				return err
			}
			if pid == procID {
				return errors.New("Process already exists")
			}
			if isProcAlive(pid) == false {
				removeProc(pid, db)
			}
		}

		if err != nil {
			return err
		}
	}

	strdata := string(data)
	strdata = strdata + ":" + strconv.Itoa(procID)

	err = db.Put(wo, []byte("procs"), []byte(strdata))
	return err
}

func ListProcs(db *levigo.DB) (map[int]int, error) {
	ro := levigo.NewReadOptions()
	procs, err := db.Get(ro, []byte("procs"))

	if err != nil {
		return map[int]int{}, err
	}

	procs_arr := strings.Split(string(procs), ":")
	var procs_arr2 = map[int]int{}
	for _, p := range procs_arr {
		status, err := db.Get(ro, []byte("status:"+p))
		if err == nil {
			p_int, _ := strconv.Atoi(p)
			procs_arr2[p_int], _ = strconv.Atoi(string(status))
		}
	}
	return procs_arr2, err

}
