package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cagnosolutions/adb"
	"github.com/cagnosolutions/mg"
	"github.com/cagnosolutions/web"
)

// global vars
var tc *web.TmplCache
var mx *web.Mux
var db *adb.DB = adb.NewDB()

var MG_DOMAIN = "api.mailgun.net/v3/sandbox73d66ccb60f948708fcaf2e2d1b3cd4c.mailgun.org"
var MG_KEY = "key-173701b40541299bd3b7d40c3ac6fd43"

func init() {
	db.AddStore("employee")
	db.AddStore("customer")
	db.AddStore("note")
	db.AddStore("task")

	web.SESSDUR = time.Minute * 60
	web.AMANAGER = true
	mx = web.NewMux()

	// unsecure routes
	mx.AddRoutes(login, loginPost, logout, makeUsers)

	// main page
	mx.AddSecureRoutes(EMPLOYEE, index)

	// employee management routes
	mx.AddSecureRoutes(ADMIN, employeeAll, employeeView, employeeSave, employeeDel, adminEmployeeTask, adminEmployeeTaskAll, adminCustomerTask, adminCustomerTaskAll)
	mx.AddSecureRoutes(ADMIN, adminTask, adminTasksave, adminTaskAll)

	mx.AddSecureRoutes(EMPLOYEE, saveHomePage, account, task, taskAll, taskMarkStart, taskMarkComplete, taskMarkNote)

	// customer management routes
	mx.AddSecureRoutes(EMPLOYEE, customerAll, customerView, customerSave, customerNoteSave)
	mx.AddSecureRoutes(ADMIN, customerDel, customerAllExport, customerAllExportDownload)

	// update session
	mx.AddSecureRoutes(ALL, updateSession, collapse)

	web.Funcs["lower"] = strings.ToLower
	web.Funcs["size"] = PrettySize
	web.Funcs["formatDate"] = FormatDate
	web.Funcs["toJson"] = ToJson
	web.Funcs["toBase64Json"] = ToBase64Json
	web.Funcs["title"] = strings.Title
	web.Funcs["idTime"] = IdTime
	web.Funcs["add"] = add
	web.Funcs["prettyDate"] = PrettyDate
	web.Funcs["prettyDateTime"] = PrettyDateTime
	web.Funcs["queryEscape"] = url.QueryEscape
	tc = web.NewTmplCache()

	defaultUsers()

	mg.SetCredentials(MG_DOMAIN, MG_KEY)

}

// main http listener
func main() {
	fmt.Println("DID YOU REGISTER ANY NEW ROUTES?")
	log.Fatal(http.ListenAndServe(":8080", mx))
}

var updateSession = web.Route{"POST", "/updateSession", func(w http.ResponseWriter, r *http.Request) {
	return
}}

var collapse = web.Route{"GET", "/collapse", func(w http.ResponseWriter, r *http.Request) {
	if web.GetSess(r, "collapse").(bool) {
		web.PutSess(w, r, "collapse", false)
	} else {
		web.PutSess(w, r, "collapse", true)
	}
	ajaxResponse(w, `{"error":false}`)
	return
}}
