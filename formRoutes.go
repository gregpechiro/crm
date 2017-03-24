package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cagnosolutions/web"
)

var driverFormAdd = web.Route{"POST", "/driver/document", func(w http.ResponseWriter, r *http.Request) {
	driverId := r.FormValue("id")
	redirect := r.FormValue("redirect")
	var driver Driver
	if !db.Get("driver", driverId, &driver) {
		web.SetErrorRedirect(w, r, redirect, "Error adding documents")
		return
	}
	docIds := strings.Split(r.FormValue("docIds"), ",")
	for _, docId := range docIds {
		id := strconv.Itoa(int(time.Now().UnixNano()))
		doc := Document{
			Id:         id,
			Name:       "dqf-" + docId,
			DocumentId: "dqf-" + docId,
			Complete:   false,
			CompanyId:  driver.CompanyId,
			DriverId:   driver.Id,
		}
		db.Add("document", id, doc)
	}

	web.SetSuccessRedirect(w, r, redirect, "Successfully added forms")
	return
}}

var formView = web.Route{"GET", "/document/:id", func(w http.ResponseWriter, r *http.Request) {
	var document Document
	var driver Driver
	var company Company
	var vehicles []Vehicle
	ok := db.Get("document", r.FormValue(":id"), &document)
	if !ok {
		web.SetErrorRedirect(w, r, "/", "Error, retrieving document.")
		return
	}
	db.Get("driver", document.DriverId, &driver)
	db.Get("company", document.CompanyId, &company)
	for _, vId := range document.VehicleIds {
		var vehicle Vehicle
		db.Get("vehicle", vId, &vehicle)
		vehicles = append(vehicles, vehicle)
	}

	tc.Render(w, r, document.DocumentId+".tmpl", web.Model{
		"document": document,
		"driver":   driver,
		"company":  company,
		"vehicles": vehicles,
	})
	return
}}

var formSave = web.Route{"POST", "/document/save", func(w http.ResponseWriter, r *http.Request) {
	var document Document
	db.Get("document", r.FormValue("id"), &document)
	document.Data = r.FormValue("data")
	db.Set("document", document.Id, document)
	web.SetFlash(w, "alertSuccess", "Successfully saved form")
	ajaxResponse(w, `{"status":"success","msg":"Successfully saved document", "redirect":"`+r.FormValue("redirect")+`"}`)
	return
}}

var formComplete = web.Route{"POST", "/document/complete", func(w http.ResponseWriter, r *http.Request) {
	var document Document
	db.Get("document", r.FormValue("id"), &document)
	document.Data = r.FormValue("data")
	document.Complete = true
	db.Set("document", document.Id, document)
	web.SetFlash(w, "alertSuccess", "Successfully completed form")
	ajaxResponse(w, `{"status":"success","msg":"Successfully saved document", "redirect":"`+r.FormValue("redirect")+`"}`)
	return
}}

var formDel = web.Route{"POST", "/document/del/:driverId/:docId", func(w http.ResponseWriter, r *http.Request) {
	db.Del("document", r.FormValue(":docId"))
	web.SetSuccessRedirect(w, r, r.FormValue("redirect"), "Successfully deleted form")
	return
}}
