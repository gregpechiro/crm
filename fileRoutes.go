package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/cagnosolutions/web"
)

/* --- Driver File Management --- */
var driverFileUpload = web.Route{"POST", "/driver/upload", func(w http.ResponseWriter, r *http.Request) {
	driverId := r.FormValue("id")
	if driverId == "" {
		log.Printf("main.go -> uploadDriverFile() -> no driver id specified")
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file"}`)
		return
	}
	path := "upload/driver/" + driverId + "/"
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("main.go -> uploadDriverFile() -> os.MkdirAll() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file"}`)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("main.go -> uploadDriverFile() -> r.FormFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file `+handler.Filename+`"}`)
		return
	}
	defer file.Close()

	f, err := os.OpenFile(path+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("main.go -> uploadDriverFile() -> os.OpenFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file `+handler.Filename+`"}`)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	ajaxResponse(w, `{"error":false,"msg":"Successfully uploaded file `+handler.Filename+`"}`)
	return
}}

var driverFileView = web.Route{"GET", "/driver/file/:id/:file", func(w http.ResponseWriter, r *http.Request) {
	server := http.StripPrefix("/driver/file/", http.FileServer(http.Dir("upload/driver/")))
	server.ServeHTTP(w, r)
	return
}}

var driverFileDel = web.Route{"POST", "/driver/file/:id/:file", func(w http.ResponseWriter, r *http.Request) {
	driverId := r.FormValue(":id")
	filename := r.FormValue(":file")
	if err := os.Remove("upload/driver/" + driverId + "/" + filename); err != nil {
		web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Error deleting file")
		return
	}
	web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Successfully deleted file")
	return
}}

/* --- Company File Management --- */

var companyFileApi = web.Route{"GET", "/company/:id/all-files", func(w http.ResponseWriter, r *http.Request) {
	companyId := r.FormValue(":id")
	path, _ := url.QueryUnescape(r.FormValue("path"))
	files, err := ioutil.ReadDir("./upload/company/" + companyId + "/files/" + path)
	if err != nil {
		ajaxResponse(w, `{"id":"#","children":false}`)
		return
	}
	// var nodes []Node
	nodes := []Node{}
	for _, file := range files {
		if file.Name()[0] != '.' {
			n := Node{}
			p := path + "/" + file.Name()
			n.Id = url.QueryEscape(p)
			n.Text = file.Name()
			n.Type = "file"
			if file.IsDir() {
				n.Type = "dir"
				n.Children = !IsEmptyDir("./upload/company/" + companyId + "/files/" + p)
				n.State = "closed"
			}
			nodes = append(nodes, n)
		}
	}
	b, err := json.Marshal(nodes)
	respString := string(b)
	if err != nil {
		respString = `{"id":"#","Children":false}`
	}
	ajaxResponse(w, respString)
	return
}}

var companyFileUpload = web.Route{"POST", "/company/:id/upload", func(w http.ResponseWriter, r *http.Request) {
	companyId := r.FormValue(":id")
	if companyId == "" {
		log.Printf("main.go -> uploadCompanyFile() -> os.MkdirAll() -> no company id specified")
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file"}`)
		return
	}

	p, _ := url.QueryUnescape(r.FormValue("path"))
	path := "upload/company/" + companyId + "/files/" + p + "/"

	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Println("MKdirAll:", path)
		log.Printf("main.go -> uploadCompanyFile() -> os.MkdirAll() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file"}`)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("main.go -> uploadCompanyFile() -> r.FormFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file `+handler.Filename+`"}`)
		return
	}
	defer file.Close()
	f, err := os.OpenFile(path+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("main.go -> uploadCompanyFile() -> os.OpenFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file `+handler.Filename+`"}`)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	ajaxResponse(w, `{"error":false,"msg":"Successfully uploaded file `+handler.Filename+`"}`)
	return
}}

var companyFileView = web.Route{"GET", "/company/:id/file", func(w http.ResponseWriter, r *http.Request) {
	path := "upload/company/" + r.FormValue(":id") + "/files" + r.FormValue("path")
	http.ServeFile(w, r, path)
	return
}}

var companyFolderNew = web.Route{"POST", "/company/:id/mkdir", func(w http.ResponseWriter, r *http.Request) {
	if r.FormValue(":id") == "" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error deleting file/folder")
		return
	}
	if r.FormValue("folder") == "" || r.FormValue("path") == "" || r.FormValue("path") == "#" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error creating new folder")
		return
	}
	p, _ := url.QueryUnescape(r.FormValue("path"))
	path := "upload/company/" + r.FormValue(":id") + "/files" + p
	if err := os.MkdirAll(path+"/"+r.FormValue("folder"), 0755); err != nil {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error creating new folder")
		return
	}
	web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Successfully created new folder")
	return
}}

var companyFileDel = web.Route{"POST", "/company/:id/file/del", func(w http.ResponseWriter, r *http.Request) {
	companyId := r.FormValue(":id")
	if companyId == "" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error deleting file/folder")
		return
	}
	if r.FormValue("path") == "" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error deleting file/folder")
		return
	}
	p, _ := url.QueryUnescape(r.FormValue("path"))
	path := "upload/company/" + companyId + "/files" + p
	switch path {
	case "upload/company/" + companyId + "/files/public":
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "You cannot delete the public folder")
		return
	case "upload/company/" + companyId + "/files/private":
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "You cannot delete the private folder")
		return
	}
	if err := os.RemoveAll(path); err != nil {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error deleting file/folder")
		return
	}
	web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Successfully deleted file/folder")
	return
}}

var companyFileMove = web.Route{"POST", "/company/:id/file/move", func(w http.ResponseWriter, r *http.Request) {
	companyId := r.FormValue(":id")
	if companyId == "" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error "+r.FormValue("type")+"ing file/folder")
		return
	}
	path := "upload/company/" + companyId + "/files"
	from, _ := url.QueryUnescape(r.FormValue("from"))
	to, _ := url.QueryUnescape(r.FormValue("to"))
	if from == "" || to == "" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error "+r.FormValue("type")+"ing file/folder")
		return
	}
	if path+from == "upload/company/"+companyId+"/files/public" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error "+r.FormValue("type")+"ing file/folder<br>Cannot "+r.FormValue("type")+"e public folder")
		return
	}
	if path+from == "upload/company/"+companyId+"/files/private" {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error "+r.FormValue("type")+"ing file/folder<br>Cannot "+r.FormValue("type")+"e private folder")
		return
	}

	if err := os.Rename(path+from, path+to); err != nil {
		web.SetErrorRedirect(w, r, r.FormValue("redirect"), "Error "+r.FormValue("type")+"ing file")
		return
	}
	web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Successfully "+r.FormValue("type")+"ed file")
	return
}}

/* --- Vehicle File Management --- */
var vehicleFileUpload = web.Route{"POST", "/vehicle/upload", func(w http.ResponseWriter, r *http.Request) {
	vehicleId := r.FormValue("id")
	if vehicleId == "" {
		log.Printf("main.go -> uploadVehicleFile() -> os.MkdirAll() -> no vehicle id specified")
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file"}`)
		return
	}
	path := "upload/vehicle/" + vehicleId + "/"
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("main.go -> uploadVehicleFile() -> os.MkdirAll() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file"}`)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("main.go -> uploadVehicleFile() -> r.FormFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file `+handler.Filename+`"}`)
		return
	}
	defer file.Close()
	f, err := os.OpenFile(path+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("main.go -> uploadVehicleFile() -> os.OpenFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error uploading file `+handler.Filename+`"}`)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	ajaxResponse(w, `{"error":false,"msg":"Successfully uploaded file `+handler.Filename+`"}`)
	return
}}

var vehicleFileView = web.Route{"GET", "/vehicle/file/:id/:file", func(w http.ResponseWriter, r *http.Request) {
	server := http.StripPrefix("/vehicle/file/", http.FileServer(http.Dir("upload/vehicle/")))
	server.ServeHTTP(w, r)
	return
}}

var vehicleFileDel = web.Route{"POST", "/vehicle/file/:id/:file", func(w http.ResponseWriter, r *http.Request) {
	vehicleId := r.FormValue(":id")
	filename := r.FormValue(":file")
	if err := os.Remove("upload/vehicle/" + vehicleId + "/" + filename); err != nil {
		web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Error deleting file")
		return
	}
	web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Successfully deleted file")
	return
}}
