package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jmhodges/levigo"
	"os"
	"strconv"
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

	return db, nil
}

func isProcAlive(pid int) bool {
	_, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
        return true
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
