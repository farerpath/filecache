/*
 * FileCache Code based from filequeue
 */

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/farerpath/randstr"
	"github.com/urfave/negroni"
)

const FILE_DIR_NAME = "file"

var FILES map[string]string

func main() {
	FILES = make(map[string]string)

	port := flag.String("port", "80", "bind port (default: 80)")

	handler := http.HandlerFunc(requestHandler)

	n := negroni.New(negroni.NewLogger(), negroni.NewRecovery())
	n.UseHandler(handler)

	err := http.ListenAndServe(":"+*port, n)
	log.Fatal(err)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Download File
		fileID := r.FormValue("file_id")

		path := makeSavePath(fileID, FILES[fileID])

		if _, err := os.Stat(path); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, path)

	} else if r.Method == http.MethodPost {
		// Queue PUT request

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fileBuffer, err := ioutil.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fileID := randstr.GenerateRandomString(8)

		path := makeSavePath(fileID, fileHeader.Filename)

		os.MkdirAll("file/"+fileID, os.ModePerm)       // Create dir
		err = ioutil.WriteFile(path, fileBuffer, 0644) // Write file
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Print(err)
			return
		}

		FILES[fileID] = fileHeader.Filename

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fileID))
	} else if r.Method == http.MethodDelete {
		fileID := r.FormValue("file_id")

		path := makeSavePath(fileID, FILES[fileID])

		if _, err := os.Stat(path); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err := os.RemoveAll(FILE_DIR_NAME + "/" + fileID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func makeSavePath(fileID string, fileName string) string {
	return FILE_DIR_NAME + "/" + fileID + "/" + fileName
}
