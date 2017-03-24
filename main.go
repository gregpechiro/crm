package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
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
	db.AddStore("company")
	db.AddStore("company-service")
	db.AddStore("company-service-emails")
	db.AddStore("driver")
	db.AddStore("vehicle")
	db.AddStore("document")
	db.AddStore("event")
	db.AddStore("note")
	db.AddStore("comment")
	db.AddStore("emailTemplate")
	db.AddStore("scheduled-email")
	db.AddStore("grouped-email")
	db.AddStore("notification")
	db.AddStore("company-features")
	db.AddStore("meta")
	db.AddStore("violation-cache")
	db.AddStore("safer-cache")
	db.AddStore("task")

	web.DEFAULT_ERR_ROUTE = web.Route{"GET", "/error/:code", func(w http.ResponseWriter, r *http.Request) {
		code, err := strconv.Atoi(r.FormValue(":code"))
		if err != nil {
			code = 500
		}
		var page = ""
		switch web.GetRole(r) {
		case "DEVELOPER", "ADMIN":
			page = HTTP_ERROR_ADMIN
		case "EMPLOYEE":
			page = HTTP_ERROR_EMPLOYEE
		case "COMPANY":
			page = HTTP_ERROR_CUSTOMER
		default:
			page = HTTP_ERROR_DEFAULT
		}
		w.Header().Set("Content-Type", "text/html; utf-8")
		fmt.Fprintf(w, page, code, http.StatusText(int(code)), code)
		return
	}}

	web.SESSDUR = time.Minute * 60
	web.AMANAGER = true
	mx = web.NewMux()

	// unsecure routes
	mx.AddRoutes(login, loginPost, logout)

	// main page
	mx.AddSecureRoutes(EMPLOYEE, index)

	// email routes
	mx.AddSecureRoutes(ADMIN, emailTemplateAll, emailTemplateView, emailTemplateSave, emailTest, emailTestSend)

	// employee management routes
	mx.AddSecureRoutes(ADMIN, employeeAll, employeeView, employeeSave, employeeDel, adminEmployeeTask, adminEmployeeTaskAll, adminCompanyTask, adminCompanyTaskAll)
	mx.AddSecureRoutes(ADMIN, adminTask, adminTasksave, adminTaskAll)

	mx.AddSecureRoutes(EMPLOYEE, saveHomePage, cnsHome, cnsAction, cnsTask, cnsTaskAll, cnsTaskMarkStart, cnsTaskMarkComplete, cnsTaskMarkNote)

	// company management routes
	mx.AddSecureRoutes(EMPLOYEE, companyAll, companyView, companySave, companyNoteSave, companyServiceView, companyServiceSave, companyServiceNotify, companyPasswordReset)
	mx.AddSecureRoutes(EMPLOYEE, companyFormAll, companyFormAdd, companyFormDel, companyFormArchive, companyFileAll)
	mx.AddSecureRoutes(EMPLOYEE, companyNotificationAll, companyNotificationAdd, companyNotificaltionDel, companyFeature, companyFeatureSave, companyViolation, companySafer)
	mx.AddSecureRoutes(EMPLOYEE, companyGlobalNotifyLastSet, companyGlobalNotifySet, companyGlobalNotifyLastReset, companyGlobalNotifyReset)
	mx.AddSecureRoutes(ADMIN, companyDel, companyAllExport, companyAllExportDownload)

	mx.AddSecureRoutes(ALL, companyFileApi, companyFileUpload, companyFileView, companyFolderNew, companyFileDel, companyFileMove)

	// company vehicle management routes
	mx.AddSecureRoutes(EMPLOYEE, companyVehicleAll, companyVehicleView, companyVehicleSave, companyVehicleFile)
	mx.AddSecureRoutes(ALL, vehicleFileUpload, vehicleFileView, vehicleFileDel)

	// driver management routes
	mx.AddSecureRoutes(EMPLOYEE, companyDriverAll, companyDriverImport, companyDriverImportUpload, companyDriverImportConvert, companyDriverView, companyDriverSave, companyDriverFileAll, companyDriverFormAll, companyDriverDel, companyDriverTransfer)
	mx.AddSecureRoutes(EMPLOYEE, driverGlobalNotifyLastSet, driverGlobalNotifySet)
	mx.AddSecureRoutes(ALL, driverFileUpload, driverFileView, driverFileDel, driverFormAdd)

	// document management routes
	mx.AddSecureRoutes(ALL, formView, formSave, formComplete, formDel)

	// update session
	mx.AddSecureRoutes(ALL, updateSession, collapse)

	// development routes
	mx.AddSecureRoutes(DEVELOPER, devComments, stats)
	mx.AddRoutes(makeUsers, GetComment, PostComent)
	mx.AddRoutes(httpError)

	//customer routes
	mx.AddRoutes(customerLogin, customerLoginPost, customerLogout)
	mx.AddSecureRoutes(COMPANY, customerHome, customerInfo, customerDriver, customerVehicle, customerForm, customerPasswordSave)
	mx.AddSecureRoutes(COMPANY, customerVehicleView, customerDriverView, customerDriverForm, customerDriverFile)
	mx.AddSecureRoutes(COMPANY, customerFile, customerVehicleFile, customerViolation, customerSafer)

	mx.AddSecureRoutes(ALL, customerViolationRestSave, customerSaferRestSave)

	// misc routes
	mx.AddSecureRoutes(ALL, ajaxProxy)

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
	go SendToday(6)
	fmt.Println("DID YOU REGISTER ANY NEW ROUTES?")
	log.Fatal(http.ListenAndServe(":8080", mx))
}

var testDB = web.Route{"GET", "/test/db", func(w http.ResponseWriter, r *http.Request) {
	testEmployees()
	testCompanies()
	testDrivers()
	web.SetMsgRedirect(w, r, "/", "Please check terminal for results")
	return
}}

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

var ajaxProxy = web.Route{"POST", "/ajax/proxy", func(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	if path == "" {
		ajaxResponse(w, `{"error": true}`)
		return
	}
	resp, err := http.Get(path)
	if err != nil {
		ajaxResponse(w, `{"error": true}`)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ajaxResponse(w, `{"error": true}`)
		return
	}

	ajaxResponse(w, `{"error": false,"data":"`+base64.StdEncoding.EncodeToString(b)+`"}`)
	return
}}

// comment section for development use only
var GetComment = web.Route{"GET", "/comment", func(w http.ResponseWriter, r *http.Request) {
	tc.Render(w, r, "comment.tmpl", web.Model{
		"return":  r.FormValue("return"),
		"comment": true,
		"page":    r.FormValue("page"),
	})
	return
}}

var PostComent = web.Route{"POST", "/comment", func(w http.ResponseWriter, r *http.Request) {
	id := strconv.Itoa(int(time.Now().UnixNano()))
	var comment Comment
	r.ParseForm()
	FormToStruct(&comment, r.Form, "")
	comment.Id = id
	db.Set("comment", id, comment)
	web.SetSuccessRedirect(w, r, comment.Url, "Successfully added comment")
	return
}}
