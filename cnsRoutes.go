package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cagnosolutions/adb"
	"github.com/cagnosolutions/mg"
	"github.com/cagnosolutions/web"
)

var logout = web.Route{"GET", "/logout", func(w http.ResponseWriter, r *http.Request) {
	web.Logout(w)
	http.Redirect(w, r, "/login", 303)
	return
}}

var login = web.Route{"GET", "/login", func(w http.ResponseWriter, r *http.Request) {
	tc.Render(w, r, "login.tmpl", web.Model{})
	return
}}

var loginPost = web.Route{"POST", "/login", func(w http.ResponseWriter, r *http.Request) {
	email, pass := r.FormValue("email"), r.FormValue("password")
	var employee Employee
	if !db.Auth("employee", email, pass, &employee) {
		web.SetErrorRedirect(w, r, "/login", "Incorrect username or password")
		return
	}
	sess := web.Login(w, r, employee.Role)
	sess.PutId(w, employee.Id)
	sess["email"] = employee.Email
	sess["collapse"] = false
	web.PutMultiSess(w, r, sess)
	redirect := "/cns"
	if employee.Home != "" {
		redirect = employee.Home
	}
	web.SetSuccessRedirect(w, r, redirect, "Welcome "+employee.FirstName)
	return
}}

var index = web.Route{"GET", "/", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/cns", 303)
	return
}}

var cnsHome = web.Route{"GET", "/cns", func(w http.ResponseWriter, r *http.Request) {
	employeeId := web.GetId(r)
	var employee Employee
	if !db.Get("employee", employeeId, &employee) {
		web.Logout(w)
		web.SetErrorRedirect(w, r, "/login", "Error finding your account")
		return
	}
	tc.Render(w, r, "cns-home.tmpl", web.Model{
		"employee": employee,
	})
}}

var cnsAction = web.Route{"GET", "/cns/action", func(w http.ResponseWriter, r *http.Request) {
	tc.Render(w, r, "cns-action.tmpl", nil)
}}

var cnsTask = web.Route{"GET", "/cns/task", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/cns/task/today", 303)
	return
}}

var cnsTaskAll = web.Route{"GET", "/cns/task/:page", func(w http.ResponseWriter, r *http.Request) {
	employeeId := web.GetId(r)
	var employee Employee
	if !db.Get("employee", employeeId, &employee) {
		web.Logout(w)
		web.SetErrorRedirect(w, r, "/login", "Error finding your account")
		return
	}
	page := r.FormValue(":page")
	var tasks []Task
	switch page {
	case "overdue":
		beg, _ := Today()
		db.TestQuery("task", &tasks, adb.Eq("employeeId", `"`+employee.Id+`"`), adb.Lt("assignedTime", strconv.Itoa(int(beg))), adb.Eq("complete", "false"))
	case "incomplete":
		db.TestQuery("task", &tasks, adb.Eq("employeeId", `"`+employee.Id+`"`), adb.Eq("complete", "false"))
	case "complete":
		db.TestQuery("task", &tasks, adb.Eq("employeeId", `"`+employee.Id+`"`), adb.Eq("complete", "true"))
	default:
		page = "today"
		beg, end := Today()
		db.TestQuery("task", &tasks, adb.Eq("employeeId", `"`+employee.Id+`"`), adb.Gt("assignedTime", strconv.Itoa(int(beg))), adb.Lt("assignedTime", strconv.Itoa(int(end))))
	}
	GetTaskEmployeeView(tasks)
	tc.Render(w, r, "cns-task.tmpl", web.Model{
		"employee": employee,
		"tasks":    tasks,
		"page":     page,
	})
}}

var cnsTaskMarkStart = web.Route{"POST", "/cns/task/:id/start", func(w http.ResponseWriter, r *http.Request) {
	var task Task
	db.Get("task", r.FormValue(":id"), &task)
	task.StartedTime = time.Now().Unix()
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, "/cns/task/today", "Successfully Started Task")
	return
}}

var cnsTaskMarkComplete = web.Route{"POST", "/cns/task/:id/complete", func(w http.ResponseWriter, r *http.Request) {
	var task Task
	db.Get("task", r.FormValue(":id"), &task)
	task.CompletedTime = time.Now().Unix()
	task.Complete = true
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, "/cns/task/today", "Successfully Started Task")
	return
}}

var cnsTaskMarkNote = web.Route{"POST", "/cns/task/:id/note", func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in controller")
	var task Task
	db.Get("task", r.FormValue(":id"), &task)
	if r.FormValue("notes") == "" {
		web.SetErrorRedirect(w, r, "/cns/task/today", "Error notes field was empty")
		return
	}
	task.Notes += "<li>" + r.FormValue("notes") + "</li>"
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, "/cns/task/today", "Successfully saved notes")
	return
}}

var saveHomePage = web.Route{"POST", "/cns/employee/:id/homepage", func(w http.ResponseWriter, r *http.Request) {
	var employee Employee
	db.Get("employee", r.FormValue(":id"), &employee)
	employee.Home = r.FormValue("url")
	db.Set("employee", employee.Id, employee)
	ajaxResponse(w, `{"error":false}`)
	return
}}

var companyGlobalNotifyLastSet = web.Route{"GET", "/cns/company/global/notify", func(w http.ResponseWriter, r *http.Request) {
	var date time.Time
	resp := "The customer service notifications were never set"
	if db.Get("meta", "companyNotifySetDate", &date) {
		resp = "The customer notifications were last set on " + date.Format("1/2/2006")
	}
	ajaxResponse(w, `{"error":false,"msg":"`+resp+`"}`)
	return
}}

var companyGlobalNotifySet = web.Route{"POST", "/cns/company/global/notify", func(w http.ResponseWriter, r *http.Request) {
	var companyServices []CompanyService
	db.All("company-service", &companyServices)
	for _, service := range companyServices {
		service.GenNotifications()
		db.Set("company-service", service.Id, service)
	}
	date := time.Now()
	db.Set("meta", "companyNotifySetDate", date)
	ajaxResponse(w, `{"error":false, "msg":"Successfully added service notifications to all customers"}`)
	return
}}

var companyGlobalNotifyLastReset = web.Route{"GET", "/cns/company/global/notify/reset", func(w http.ResponseWriter, r *http.Request) {
	var date time.Time
	resp := "The customer service notifications have never been reset"
	if db.Get("meta", "companyNotifyResetDate", &date) {
		resp = "The custoer service notifications were last reset on " + date.Format("1/2/2006")
	}
	ajaxResponse(w, `{"error":false,"msg":"`+resp+`"}`)
	return
}}

var companyGlobalNotifyReset = web.Route{"POST", "/cns/company/global/notify/reset", func(w http.ResponseWriter, r *http.Request) {
	var companyServices []CompanyService
	db.All("company-service", &companyServices)
	for _, service := range companyServices {
		service.ResetNotifications()
		db.Set("company-service", service.Id, service)
	}
	var notifications []Notification
	db.TestQuery("notification", &notifications, adb.Eq("type", "COMPANY"), adb.Eq("subType", "SERVICE"))
	for _, notification := range notifications {
		db.Del("notification", notification.Id)
	}
	date := time.Now()
	db.Set("meta", "companyNotifyResetDate", date)
	ajaxResponse(w, `{"error":false, "msg":"Successfully reset all customer service notifications"}`)
	return
}}

var driverGlobalNotifyLastSet = web.Route{"GET", "/cns/driver/global/notify", func(w http.ResponseWriter, r *http.Request) {
	var date time.Time
	resp := "The driver form notifications were never set"
	if db.Get("meta", "driverNotifySetDate", &date) {
		resp = "The driver form Notifications were last set on " + date.Format("1/2/2006")
	}
	ajaxResponse(w, `{"error":false,"msg":"`+resp+`"}`)
	return
}}

var driverGlobalNotifySet = web.Route{"POST", "/cns/driver/global/notify", func(w http.ResponseWriter, r *http.Request) {
	var drivers []Driver
	db.All("driver", &drivers)
	for _, driver := range drivers {
		driver.GenNotifications()
		db.Set("driver", driver.Id, driver)
	}
	date := time.Now()
	db.Set("meta", "driverNotifySetDate", date)
	ajaxResponse(w, `{"error":false, "msg":"Successfully added service notifications to all customers"}`)
	return
}}

/* --- Company Management --- */

var companyAll = web.Route{"GET", "/cns/company", func(w http.ResponseWriter, r *http.Request) {
	var companies []Company
	db.All("company", &companies)
	tc.Render(w, r, "company-all.tmpl", web.Model{
		"companies": companies,
	})
	return
}}

var companyAllExport = web.Route{"POST", "/cns/company/all/export", func(w http.ResponseWriter, r *http.Request) {
	var companies []Company
	db.All("company", &companies)
	exporter, err := NewCSVExporter(companies)
	if err != nil {
		log.Printf("cnsRoutes.go companyAllExport >> NewCSVExporter() >> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error exporting companies"}`)
		return
	}
	path := "export/"
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("cnsRoutes.go companyAllExport >> os.MkdirAll() >> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error exporting companies"}`)
		return
	}

	path = path + time.Now().Format("2006-01-02") + "_companies.csv"
	if err := exporter.Export(path); err != nil {
		log.Printf("cnsRoutes.go companyAllExport >> exporter.Export() >> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error exporting companies"}`)
		return
	}
	ajaxResponse(w, `{"error":false,"path":"/`+path+`"}`)
	return
}}

var companyAllExportDownload = web.Route{"GET", "/export/:name", func(w http.ResponseWriter, r *http.Request) {
	server := http.StripPrefix("/export", http.FileServer(http.Dir("export/")))
	server.ServeHTTP(w, r)
	return
}}

var companyView = web.Route{"GET", "/cns/company/:id", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	compId := r.FormValue(":id")
	if compId != "new" && !db.Get("company", compId, &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var notes NoteSort
	var employees []Employee
	db.TestQuery("note", &notes, adb.Eq("companyId", `"`+company.Id+`"`))
	sort.Stable(sort.Reverse(notes))
	db.All("employee", &employees)
	tc.Render(w, r, "company.tmpl", web.Model{
		"company":    company,
		"notes":      notes,
		"employees":  employees,
		"quickNotes": quickNotes,
		"employeeId": web.GetId(r),
	})
	return
}}

var companySave = web.Route{"POST", "/cns/company", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	db.Get("company", r.FormValue("id"), &company)
	FormToStruct(&company, r.Form, "")
	var companies []Company
	db.TestQuery("company", &companies, adb.Eq("email", company.Email), adb.Ne("id", `"`+company.Id+`"`))
	if len(companies) > 0 {
		end := "/new"
		if r.FormValue("id") != "" {
			end = "/" + r.FormValue("id")
		}
		web.SetErrorRedirect(w, r, "/cns/company"+end, "Error saving company. Email is already registered")
		return
	}
	if company.Id == "" {
		company.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}
	if company.SameAddress {
		company.MailingAddress = company.PhysicalAddress
	}
	if company.CreditCard.ExpirationMonth > 0 && company.CreditCard.ExpirationYear > 0 {
		company.CreditCard.ExpirationDate = strconv.Itoa(company.CreditCard.ExpirationMonth) + "/" + strconv.Itoa(company.CreditCard.ExpirationYear)
	}
	db.Set("company", company.Id, company)
	if r.FormValue("from") == "vehicle" {
		web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/vehicle", "Successfully updated insurance information")
		return
	}
	if r.FormValue("from") == "service" {
		UpdateCompanyEmails(&company)
		web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/service", "Successfully updated service information")
		return
	}
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id, "Successfully saved company")
	return
}}

var companyNoteSave = web.Route{"POST", "/cns/company/:id/note", func(w http.ResponseWriter, r *http.Request) {
	var note Note
	r.ParseForm()
	FormToStruct(&note, r.Form, "")
	if note.Id == "" {
		note.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}
	dt, err := time.Parse("01/02/2006 3:04 PM", r.FormValue("dateTime"))
	if err != nil {
		log.Printf("cnsRoutes.go >> companySaveNotes >> time.Parse() >> %v\n", err)
	}
	note.StartTime = dt.Unix()
	note.EndTime = dt.Unix()
	note.StartTimePretty = r.FormValue("dateTime")
	note.EndTimePretty = r.FormValue("dateTime")
	db.Set("note", note.Id, note)
	web.SetSuccessRedirect(w, r, "/cns/company/"+r.FormValue(":id"), "Successfully saved note")
	return
}}

var companyServiceView = web.Route{"GET", "/cns/company/:id/service", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var companyService CompanyService
	db.TestQueryOne("company-service", &companyService, adb.Eq("companyId", `"`+company.Id+`"`))
	tc.Render(w, r, "company-service.tmpl", web.Model{
		"company":        company,
		"companyService": companyService,
	})
	return
}}

var companyServiceSave = web.Route{"POST", "/cns/company/:id/service", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var companyService CompanyService
	db.Get("company-service", r.FormValue("id"), &companyService)
	FormToStruct(&companyService, r.Form, "")
	if companyService.CompanyId != company.Id {
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/service", "Error saving company services")
		return
	}
	if companyService.Id == "" {
		companyService.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}
	db.Set("company-service", companyService.Id, companyService)
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/service", "Successfully saved company Services")

}}

var companyServiceNotify = web.Route{"POST", "/cns/company/:id/service", func(w http.ResponseWriter, r *http.Request) {
	compId := r.FormValue(":id")
	var companyService CompanyService
	db.TestQueryOne("company-service", &companyService, adb.Eq("companyId", `"`+compId+`"`))
	companyService.GenNotifications()
	db.Set("company-service", companyService.Id, companyService)
	web.SetSuccessRedirect(w, r, "/cns/company/"+compId+"/notification", "Successfully created notifications")
	return
}}

var companyFormAll = web.Route{"GET", "/cns/company/:id/form", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var docs []Document
	db.TestQuery("document", &docs, adb.Eq("companyId", `"`+company.Id+`"`), adb.Eq("stateForm", "true"))

	var vehicles []Vehicle
	db.TestQuery("vehicle", &vehicles, adb.Eq("companyId", `"`+company.Id+`"`))

	tc.Render(w, r, "company-form.tmpl", web.Model{
		"company":  company,
		"docs":     docs,
		"forms":    CompanyForms,
		"vehicles": vehicles,
	})
	return
}}

var companyFormAdd = web.Route{"POST", "/cns/company/:id/form", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company/", "Error finding company")
		return
	}

	id := strconv.Itoa(int(time.Now().UnixNano()))
	docId := r.FormValue("name")
	var vehicleIds []string
	if r.FormValue("vehicleIds") != "" {
		vehicleIds = strings.Split(r.FormValue("vehicleIds"), ",")
	}
	doc := Document{
		Id:         id,
		Name:       docId,
		DocumentId: "st-" + strings.ToLower(strings.Replace(docId, " ", "_", -1)),
		Complete:   false,
		CompanyId:  company.Id,
		VehicleIds: vehicleIds,
		StateForm:  true,
	}
	db.Add("document", id, doc)
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/form", "Successfully added forms")
	return
}}

var companyFormDel = web.Route{"POST", "/cns/company/:compId/form/:formId", func(w http.ResponseWriter, r *http.Request) {
	var form Document
	compId := r.FormValue(":compId")
	if !db.Get("document", r.FormValue(":formId"), &form) || form.CompanyId != compId {
		web.SetErrorRedirect(w, r, "/cns/company/"+compId+"/form", "Error deleting from")
		return
	}
	db.Del("document", form.Id)
	web.SetSuccessRedirect(w, r, "/cns/company/"+compId+"/form", "Successfully deleted form")
	return

}}

var companyFormArchive = web.Route{"POST", "/cns/company/:id/archive", func(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("name") == "" {
		fmt.Println("no name")
		ajaxResponse(w, `{"error":true,"msg":"Error archiving form. No name was specified.<br>Please enter a name and try again."}`)
		return
	}
	name := r.FormValue("name") + ".pdf"
	companyId := r.FormValue(":id")
	if companyId == "" {
		log.Printf("main.go -> uploadCompanyFile() -> os.MkdirAll() -> no company id specified")
		ajaxResponse(w, `{"error":true,"msg":"Error archiving form as `+name+`.<br>Please try again"}`)
		return
	}
	path := "upload/company/" + companyId + "/files/private/archived-forms/"
	if _, err := os.Stat(path + name); err == nil {
		ajaxResponse(w, `{"error":true,"msg":"Error archiving form as `+name+`. <br>A file or archived form aleady exists with that name.<br>Please rename and try again"}`)
		return
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("main.go -> companyFormArchive() -> os.MkdirAll() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error archiving form as `+name+`.<br>Please try again"}`)
		return
	}
	r.ParseForm()
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("main.go -> companyformArchive() -> r.FormFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error archiving form as `+name+`.<br>Please try again"}`)
		return
	}

	defer file.Close()
	f, err := os.OpenFile(path+name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("main.go -> companyFormArchive() -> os.OpenFile() -> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error archiving form as `+name+`.<br>Please try again"}`)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	ajaxResponse(w, `{"error":false,"msg":"Successfully archived form as `+name+`"}`)
	return
}}

var companyFileAll = web.Route{"GET", "/cns/company/:id/file", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	if _, err := os.Stat("upload/company/" + company.Id + "/files/public"); err != nil {
		if err := os.MkdirAll("upload/company/"+company.Id+"/files/public", 0755); err != nil {
			fmt.Fprintf(w, "error making dir \"upload/company/%s/files/public\"\n%v", company.Id, err)
		}
	}

	if _, err := os.Stat("upload/company/" + company.Id + "/files/private"); err != nil {
		if err := os.MkdirAll("upload/company/"+company.Id+"/files/private", 0755); err != nil {
			fmt.Fprintf(w, "error making dir \"upload/company/%s/files/private\"\n%v", company.Id, err)
		}
	}

	tc.Render(w, r, "company-file.tmpl", web.Model{
		"company": company,
	})
	return
}}

var companyNotificationAll = web.Route{"GET", "/cns/company/:id/notification", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var notifications []Notification
	db.TestQuery("notification", &notifications, adb.Eq("type", "COMPANY"), adb.Eq("modelId", `"`+company.Id+`"`))
	tc.Render(w, r, "company-notification.tmpl", web.Model{
		"company":       company,
		"notifications": notifications,
	})
}}

var companyNotificationAdd = web.Route{"POST", "/cns/company/:id/notification", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	nId := strconv.Itoa(int(time.Now().UnixNano()))
	notification := Notification{
		Id:      nId,
		ModelId: company.Id,
		Type:    "COMPANY",
		Title:   r.FormValue("title"),
		Body:    r.FormValue("body"),
		Manual:  true,
	}
	db.Add("notification", nId, notification)
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/notification", "Successfully added notification")
	return
}}

var companyNotificaltionDel = web.Route{"POST", "/cns/company/:id/notification/:notificationId", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var notification Notification
	if !db.Get("notification", r.FormValue(":notificationId"), &notification) {
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/notification", "Error finding notification")
		return
	}
	if notification.Type != "COMPANY" || notification.ModelId != company.Id {
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/notification", "Error deleting notification")
		return
	}
	db.Del("notification", r.FormValue(":notificationId"))
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/notification", "Successfully deleted notification")
	return
}}

var companyFeature = web.Route{"GET", "/cns/company/:id/feature", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+company.Id+`"`))
	tc.Render(w, r, "company-features.tmpl", web.Model{
		"company":         company,
		"companyFeatures": companyFeatures,
	})
}}

var companyFeatureSave = web.Route{"POST", "/cns/company/:id/feature", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var companyFeatures CompanyFeatures
	db.Get("company-features", r.FormValue("id"), &companyFeatures)
	FormToStruct(&companyFeatures, r.Form, "")
	if companyFeatures.Id == "" {
		companyFeatures.Id = strconv.Itoa(int(time.Now().UnixNano()))
		companyFeatures.CompanyId = company.Id
	}
	db.Set("company-features", companyFeatures.Id, companyFeatures)
	active, _ := strconv.ParseBool(r.FormValue("login"))
	if active {
		if company.Email == "" {
			company.Active = false
			db.Set("company", company.Id, company)
			web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/feature", "Customer must have a valid email address before enabling login.<br>All other features were saved")
			return
		}
		if company.Password == "" {
			company.Password = company.Email
		}
	}
	company.Active = active
	db.Set("company", company.Id, company)
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/feature", "Successfully saved features")
	return
}}

var companyViolation = web.Route{"GET", "/cns/company/:id/violation", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	tc.Render(w, r, "company-violation.tmpl", web.Model{
		"company":    company,
		"violations": GetCustomerViolations(company.Id),
	})
}}

var companySafer = web.Route{"GET", "/cns/company/:id/safer", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	tc.Render(w, r, "company-safer.tmpl", web.Model{
		"company": company,
		"safer":   GetCustomerSafer(company.Id),
	})
}}

var companyDel = web.Route{"POST", "/cns/company/:id/del", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}

	// delete all documents
	var documents []Document
	db.TestQuery("document", &documents, adb.Eq("companyId", `"`+company.Id+`"`))
	for _, document := range documents {
		db.Del("document", document.Id)
	}

	// delete companyService
	var companyService CompanyService
	if db.TestQueryOne("company-service", &companyService, adb.Eq("companyId", `"`+company.Id+`"`)) {
		db.Del("company-service", companyService.Id)
	}

	// delete companyServiceEmails
	var companyServiceEmails CompanyServiceEmails
	if db.TestQueryOne("company-service-emails", &companyServiceEmails, adb.Eq("companyId", `"`+company.Id+`"`)) {
		db.Del("company-service-emails", companyServiceEmails.Id)
	}

	// delete companyFeatures
	var companyFeatures CompanyFeatures
	if db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+company.Id+`"`)) {
		db.Del("company-features", companyFeatures.Id)
	}

	// delete all notes
	var notes []Note
	db.TestQuery("note", &notes, adb.Eq("companyId", `"`+company.Id+`"`))
	for _, note := range notes {
		db.Del("note", note.Id)
	}

	// delete all notifications
	var notifications []Notification
	db.TestQuery("notification", &notifications, adb.Eq("modelId", `"`+company.Id+`"`), adb.Eq("type", "COMPANY"))
	for _, notification := range notifications {
		db.Del("notification", notification.Id)
	}

	// delete all Drivers
	var drivers []Driver
	db.TestQuery("driver", &drivers, adb.Eq("companyId", `"`+company.Id+`"`))
	for _, driver := range drivers {
		DeleteDriver(driver.Id)
	}

	// delete all vehicles
	var vehicles []Vehicle
	db.TestQuery("vehicle", &vehicles, adb.Eq("companyId", `"`+company.Id+`"`))
	for _, vehicle := range vehicles {
		DeleteVehicle(vehicle.Id)
	}

	// delete company files
	os.RemoveAll("upload/company/" + company.Id + "/")

	// delete company
	db.Del("company", company.Id)

	web.SetSuccessRedirect(w, r, "/cns/company", "Successfully deleted customer")
	return
}}

var companyPasswordReset = web.Route{"POST", "/cns/company/:id/passwordReset", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}

	company.Password = company.Email
	db.Set("company", company.Id, company)
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id, `Successfully reset customer\'s password`)
	return
}}

/* --- Company Vehicle Management --- */

var companyVehicleAll = web.Route{"GET", "/cns/company/:id/vehicle", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var vehicles []Vehicle
	var state string
	switch r.FormValue("state") {
	case "active":
		db.TestQuery("vehicle", &vehicles, adb.Eq("companyId", `"`+company.Id+`"`), adb.Eq("active", "true"))
		state = "active"
	case "inactive":
		db.TestQuery("vehicle", &vehicles, adb.Eq("companyId", `"`+company.Id+`"`), adb.Eq("active", "false"))
		state = "inactive"
	default:
		db.TestQuery("vehicle", &vehicles, adb.Eq("companyId", `"`+company.Id+`"`))
		state = "all"
	}

	tc.Render(w, r, "company-vehicle-all.tmpl", web.Model{
		"company":  company,
		"vehicles": vehicles,
		"state":    state,
	})
	return
}}

var companyVehicleView = web.Route{"GET", "/cns/company/:compId/vehicle/:vId", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":compId"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	vehicleId := r.FormValue(":vId")
	var vehicle Vehicle
	if vehicleId != "new" && !db.Get("vehicle", vehicleId, &vehicle) {
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/vehicle", "Error finding vehicle")
		return
	}

	tc.Render(w, r, "company-vehicle.tmpl", web.Model{
		"company": company,
		"vehicle": vehicle,
	})
	return
}}

var companyVehicleSave = web.Route{"POST", "/cns/company/:compId/vehicle", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":compId"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var vehicle Vehicle
	db.Get("vehicle", r.FormValue("id"), &vehicle)
	FormToStruct(&vehicle, r.Form, "")
	if vehicle.Id == "" {
		vehicle.Id = strconv.Itoa(int(time.Now().UnixNano()))
		vehicle.CompanyId = company.Id
	}
	vehicle.PlateExpire = ""
	if vehicle.PlateExpireMonth != "" && vehicle.PlateExpireYear != "" {
		vehicle.PlateExpire = vehicle.PlateExpireMonth + "/" + vehicle.PlateExpireYear
	}
	db.Set("vehicle", vehicle.Id, vehicle)
	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/vehicle/"+vehicle.Id, "Successfully saved vehicle")
	return
}}

var companyVehicleFile = web.Route{"GET", "/cns/company/:compId/vehicle/:vId/file", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":compId"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	var vehicle Vehicle
	if !db.Get("vehicle", r.FormValue(":vId"), &vehicle) || vehicle.CompanyId != company.Id {
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/vehicle", "Error finding vehicle")
		return
	}

	var files []map[string]interface{}
	if fileInfos, err := ioutil.ReadDir("upload/vehicle/" + vehicle.Id); err == nil {
		for _, fileInfo := range fileInfos {
			var info = make(map[string]interface{})
			info["name"] = fileInfo.Name()
			info["size"] = fileInfo.Size()
			files = append(files, info)
		}
	}
	tc.Render(w, r, "company-vehicle-file.tmpl", web.Model{
		"company": company,
		"vehicle": vehicle,
		"files":   files,
	})
	return
}}

/* --- Company Driver Management --- */

var companyDriverAll = web.Route{"GET", "/cns/company/:id/driver", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	var drivers []Driver
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	db.TestQuery("driver", &drivers, adb.Eq("companyId", `"`+company.Id+`"`))
	tc.Render(w, r, "company-driver-all.tmpl", web.Model{
		"company": company,
		"drivers": drivers,
	})
	return
}}

var companyDriverView = web.Route{"GET", "/cns/company/:compId/driver/:id", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":compId"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	driverId := r.FormValue(":id")
	var driver Driver
	if driverId != "new" {
		if !db.Get("driver", driverId, &driver) {
			web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/driver", "Error finding driver")
			return
		}
		if driver.CompanyId != company.Id {
			web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/driver", "Error finding driver")
			return
		}
	}
	var companies []Company
	db.All("company", &companies)

	tc.Render(w, r, "company-driver.tmpl", web.Model{
		"driver":    driver,
		"company":   company,
		"companies": companies,
	})
	return
}}

var companyDriverFormAll = web.Route{"GET", "/cns/company/:compId/driver/:id/form", func(w http.ResponseWriter, r *http.Request) {
	var driver Driver
	var docs []Document
	if !db.Get("driver", r.FormValue(":id"), &driver) {
		web.SetErrorRedirect(w, r, "/cns/company/"+r.FormValue(":compId")+"/driver", "Error finding driver")
		return
	}
	if driver.CompanyId != r.FormValue(":compId") {
		web.SetErrorRedirect(w, r, "/cns/company/"+r.FormValue(":compId")+"/driver", "Error finding driver")
		return
	}
	db.TestQuery("document", &docs, adb.Eq("driverId", `"`+driver.Id+`"`), adb.Eq("companyId", `"`+driver.CompanyId+`"`))
	tc.Render(w, r, "company-driver-form.tmpl", web.Model{
		"driver": driver,
		"dqfs":   DQFS,
		"docs":   docs,
	})
	return
}}

var companyDriverFileAll = web.Route{"GET", "/cns/company/:compId/driver/:id/file", func(w http.ResponseWriter, r *http.Request) {
	var driver Driver
	if !db.Get("driver", r.FormValue(":id"), &driver) {
		web.SetErrorRedirect(w, r, "/cns/company/"+r.FormValue(":compId")+"/driver", "Error finding driver")
		return
	}
	if driver.CompanyId != r.FormValue(":compId") {
		web.SetErrorRedirect(w, r, "/cns/company/"+r.FormValue(":compId")+"/driver", "Error finding driver")
		return
	}
	var files []map[string]interface{}
	if fileInfos, err := ioutil.ReadDir("upload/driver/" + driver.Id); err == nil {
		for _, fileInfo := range fileInfos {
			var info = make(map[string]interface{})
			info["name"] = fileInfo.Name()
			info["size"] = fileInfo.Size()
			files = append(files, info)
		}
	}
	tc.Render(w, r, "company-driver-file.tmpl", web.Model{
		"driver": driver,
		"files":  files,
	})
	return
}}

var companyDriverSave = web.Route{"POST", "/cns/driver", func(w http.ResponseWriter, r *http.Request) {
	var driver Driver
	db.Get("driver", r.FormValue("id"), &driver)
	FormToStruct(&driver, r.Form, "")
	if driver.Id == "" {
		driver.Id = strconv.Itoa(int(time.Now().UnixNano()))
		driver.Password = driver.Email
		driver.Role = "DRIVER"
	}
	var company Company
	db.Get("company", driver.CompanyId, &company)
	if company.Email != "" {
		UpdateDriverEmails(&driver, company.Email)
	}
	db.Set("driver", driver.Id, driver)
	web.SetSuccessRedirect(w, r, "/cns/company/"+driver.CompanyId+"/driver/"+driver.Id, "Successfully saved driver")
	return
}}

var companyDriverDel = web.Route{"POST", "/cns/driver/:id", func(w http.ResponseWriter, r *http.Request) {
	driverId := r.FormValue(":id")
	DeleteDriver(driverId)

	web.SetSuccessRedirect(w, r, "/cns/company/"+r.FormValue("companyId")+"/driver", "Successfully deleted driver and all of the associated forms and files")
	return

}}

var companyDriverTransfer = web.Route{"POST", "/cns/driver/:id/transfer", func(w http.ResponseWriter, r *http.Request) {

	newDriverId := strconv.Itoa(int(time.Now().UnixNano()))

	driverId := r.FormValue(":id")
	var driver Driver
	db.Get("driver", driverId, &driver)

	companyId := r.FormValue("companyId")

	var documents []Document
	db.TestQuery("document", &documents, adb.Eq("driverId", `"`+driverId+`"`))
	for _, doc := range documents {
		doc.CompanyId = companyId
		db.Set("document", doc.Id, doc)

		newDocId := strconv.Itoa(int(time.Now().UnixNano()))
		doc.Id = newDocId
		doc.DriverId = newDriverId
		doc.CompanyId = driver.CompanyId
		db.Set("document", doc.Id, doc)
	}

	oldCompanyId := driver.CompanyId

	driver.CompanyId = companyId
	db.Set("driver", driverId, driver)

	driver.Id = newDriverId
	driver.Status = "TRANSFERED"
	driver.OriginalId = driverId
	driver.CompanyId = oldCompanyId
	db.Set("driver", newDriverId, driver)

	if err := CopyDir("upload/driver/"+driverId+"/", "upload/driver/"+newDriverId+"/"); err != nil {
		log.Printf("cnsRoutes.go >> transferDriver >> CopyDir() >> %v\n\n", err)
	}

	web.SetSuccessRedirect(w, r, "/cns/company/"+companyId+"/driver/"+driverId, "Successfully transfered driver")
	return
}}

var companyDriverImportUpload = web.Route{"POST", "/cns/company/:id/driver/import", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	path := "upload/company/" + company.Id + "/driver-import/"
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("main.go -> uploadCompanyFile() -> os.MkdirAll() -> %v\n", err)
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/driver", "Error uploading csv file")
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("main.go -> uploadCompanyFile() -> r.FormFile() -> %v\n", err)
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/driver", "Error uploading csv file")
		return
	}
	defer file.Close()
	f, err := os.OpenFile(path+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("main.go -> uploadCompanyFile() -> os.OpenFile() -> %v\n", err)
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/driver", "Error uploading csv file")
		return
	}
	defer f.Close()
	io.Copy(f, file)

	web.SetSuccessRedirect(w, r, "/cns/company/"+company.Id+"/driver/import/"+handler.Filename, "Successfully uploaded csv file")
	return
}}

var companyDriverImport = web.Route{"GET", "/cns/company/:id/driver/import/:file", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}
	path := "upload/company/" + company.Id + "/driver-import/" + r.FormValue(":file")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/driver", "Error finding file "+r.FormValue(":file"))
		return
	}

	file := NewCSVFile(path)

	tc.Render(w, r, "company-import.tmpl", web.Model{
		"company": company,
		"file":    r.FormValue(":file"),
		"head":    file.GetHeader(),
	})
}}

var companyDriverImportConvert = web.Route{"POST", "/cns/company/:id/driver/convert", func(w http.ResponseWriter, r *http.Request) {
	var company Company
	if !db.Get("company", r.FormValue(":id"), &company) {
		web.SetErrorRedirect(w, r, "/cns/company", "Error finding company")
		return
	}

	path := "upload/company/" + company.Id + "/driver-import/" + r.FormValue("file")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		web.SetErrorRedirect(w, r, "/cns/company/"+company.Id+"/driver", "Error finding file "+r.FormValue(":file"))
		return
	}

	r.ParseForm()
	var drivers []Driver

	csvFile := NewCSVFile(path)

	if err := csvFile.ConvertFromForm(r.Form, &drivers); err != nil {
		web.SetErrorRedirect(w, r, "/cns/company/"+r.FormValue(":id")+"/driver", err.Error())
		return
	}

	for _, driver := range drivers {
		driver.Id = strconv.Itoa(int(time.Now().UnixNano()))
		driver.CompanyId = company.Id
		db.Set("driver", driver.Id, driver)
	}

	web.SetSuccessRedirect(w, r, "/cns/company/"+r.FormValue(":id")+"/driver", "Successfully imported drivers")
	return

}}

/* --- Email Helpers --- */

func UpdateDriverEmails(driver *Driver, email string) {
	driver.LicenseExpireEmailId = updateDriverEmail(driver.LicenseExpireEmailId, "License", driver.LicenseExpire, email, -30, driver)
	driver.MedCardExpireEmailId = updateDriverEmail(driver.MedCardExpireEmailId, "Medical Card", driver.MedCardExpiry, email, -30, driver)
	driver.MVRExpireEmailId = updateDriverEmail(driver.MVRExpireEmailId, "MVR", driver.MVRExpiry, email, -30, driver)
	driver.ReviewExpireEmailId = updateDriverEmail(driver.ReviewExpireEmailId, "Review", driver.ReviewExpiry, email, -30, driver)
	driver.OneEightyExpireEmailId = updateDriverEmail(driver.OneEightyExpireEmailId, "1080", driver.OneEightyExpiry, email, -30, driver)
}

func updateDriverEmail(groupedEmailId, document, date, email string, daysAfter int, driver *Driver) string {
	emailTime, err := time.Parse("01/02/2006", date)
	if err != nil {
		return ""
	}
	emailTime = emailTime.AddDate(0, 0, daysAfter)
	emailTS := emailTime.Unix()

	var groupedEmail GroupedEmail
	if !db.Get("grouped-email", groupedEmailId, &groupedEmail) {
		groupedEmail.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}

	groupedEmail = GroupedEmail{
		DataId:    driver.Id,
		DataKey:   "drivers",
		DataStore: "driver",
		ValsKey:   driver.Id + "-" + document,
		GroupId:   "driver-formexp_" + driver.CompanyId,
	}

	groupedEmail.Vals = map[string]interface{}{
		"document": document,
		"date":     date,
	}

	if groupedEmail.Time != emailTS {
		groupedEmail.Sent = false
	}
	groupedEmail.Time = emailTS
	db.Set("grouped-email", groupedEmailId, groupedEmail)
	return groupedEmailId
}

func UpdateCompanyEmails(company *Company) {
	var serviceEmails CompanyServiceEmails
	if !db.TestQueryOne("company-service-emails", &serviceEmails, adb.Eq("companyId", `"`+company.Id+`"`)) {
		serviceEmails.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}
	var companyService CompanyService
	db.TestQueryOne("company-service", &companyService, adb.Eq("companyId", `"`+company.Id+`"`))
	if companyService.Apportion {
		serviceEmails.ApportionOneEmailId = updateCompanyEmail(serviceEmails.ApportionOneEmailId, "Apportion 1", company.Id, companyService.ApportionDateOne, company.Email, 0, 0)
		serviceEmails.ApportionTwoEmailId = updateCompanyEmail(serviceEmails.ApportionTwoEmailId, "Apportion 2", company.Id, companyService.ApportionDateTwo, company.Email, 0, 0)
	} else if serviceEmails.ApportionOneEmailId != "" {
		db.Del("scheduled-email", serviceEmails.ApportionOneEmailId)
		db.Del("scheduled-email", serviceEmails.ApportionTwoEmailId)
		serviceEmails.ApportionOneEmailId = ""
		serviceEmails.ApportionTwoEmailId = ""
	}
	if companyService.FuelTaxProgram {
		serviceEmails.FuelTaxProgramEmailId = updateCompanyEmail(serviceEmails.FuelTaxProgramEmailId, "Fuel Tax Program", company.Id, "03/30", company.Email, 3, 0)
	} else if serviceEmails.FuelTaxProgramEmailId != "" {
		db.Del("scheduled-email", serviceEmails.FuelTaxProgramEmailId)
		serviceEmails.FuelTaxProgramEmailId = ""
	}
	if companyService.FuelTaxNY {
		serviceEmails.FuelTaxNYEmailId = updateCompanyEmail(serviceEmails.FuelTaxNYEmailId, "Fuel Tax NY Program", company.Id, "03/30", company.Email, 3, 0)
	} else if serviceEmails.FuelTaxNYEmailId != "" {
		db.Del("scheduled-email", serviceEmails.FuelTaxNYEmailId)
		serviceEmails.FuelTaxNYEmailId = ""
	}
	if companyService.FuelTaxKY {
		serviceEmails.FuelTaxKYEmailId = updateCompanyEmail(serviceEmails.FuelTaxKYEmailId, "Fuel Tax KY Program", company.Id, "03/30", company.Email, 3, 0)
	} else if serviceEmails.FuelTaxNYEmailId != "" {
		db.Del("scheduled-email", serviceEmails.FuelTaxNYEmailId)
		serviceEmails.FuelTaxKYEmailId = ""
	}
	if companyService.FuelTaxNM {
		serviceEmails.FuelTaxNMEmailId = updateCompanyEmail(serviceEmails.FuelTaxNMEmailId, "Fuel Tax NM Program", company.Id, "03/30", company.Email, 3, 0)
	} else if serviceEmails.FuelTaxNMEmailId != "" {
		db.Del("scheduled-email", serviceEmails.FuelTaxNMEmailId)
		serviceEmails.FuelTaxNMEmailId = ""
	}
	if companyService.DrugConsortium {
		serviceEmails.DrugConsortiumEmailId = updateCompanyEmail(serviceEmails.DrugConsortiumEmailId, "Drug Consortium", company.Id, companyService.DrugConsortiumDate, company.Email, 0, 0)
	} else if serviceEmails.DrugConsortiumEmailId != "" {
		db.Del("scheduled-email", serviceEmails.DrugConsortiumEmailId)
		serviceEmails.DrugConsortiumEmailId = ""
	}
	if companyService.DriverFileManagement {
		serviceEmails.DriverFileManagmentEmailId = updateCompanyEmail(serviceEmails.DriverFileManagmentEmailId, "Driver File Managment", company.Id, companyService.DriverFileManagementDate, company.Email, 0, 0)
	} else if serviceEmails.DriverFileManagmentEmailId != "" {
		db.Del("scheduled-email", serviceEmails.DriverFileManagmentEmailId)
		serviceEmails.DriverFileManagmentEmailId = ""
	}
	if companyService.DOTUpdate {
		serviceEmails.DOTUpdateEmailId = updateCompanyEmail(serviceEmails.DOTUpdateEmailId, "DOT Update", company.Id, companyService.DOTUpdateDate, company.Email, 0, 0)
	} else if serviceEmails.DOTUpdateEmailId != "" {
		db.Del("scheduled-email", serviceEmails.DOTUpdateEmailId)
		serviceEmails.DOTUpdateEmailId = ""
	}
	if companyService.TwentyTwoNinety {
		serviceEmails.TwentyTwoNinetyEmailId = updateCompanyEmail(serviceEmails.TwentyTwoNinetyEmailId, "2290", company.Id, "06/01", company.Email, 0, 1)
	} else if serviceEmails.TwentyTwoNinetyEmailId != "" {
		db.Del("scheduled-email", serviceEmails.TwentyTwoNinetyEmailId)
		serviceEmails.TwentyTwoNinetyEmailId = ""
	}
	if companyService.UCR {
		serviceEmails.UCREmailId = updateCompanyEmail(serviceEmails.UCREmailId, "UCR", company.Id, "10/01", company.Email, 0, 1)
	} else if serviceEmails.UCREmailId != "" {
		db.Del("scheduled-email", serviceEmails.UCREmailId)
		serviceEmails.UCREmailId = ""
	}
	if companyService.LogAuditing {
		serviceEmails.LogAuditingEmailId = updateCompanyEmail(serviceEmails.LogAuditingEmailId, "Log Auditing", company.Id, "01/07", company.Email, 1, 0)
	} else if serviceEmails.LogAuditingEmailId != "" {
		db.Del("scheduled-email", serviceEmails.LogAuditingEmailId)
		serviceEmails.LogAuditingEmailId = ""
	}
	if companyService.CSAService {
		serviceEmails.CSAServiceEmailId = updateCompanyEmail(serviceEmails.CSAServiceEmailId, "CSA Service", company.Id, companyService.CSAServiceDate, company.Email, 0, 0)
	} else if serviceEmails.CSAServiceEmailId != "" {
		db.Del("scheduled-email", serviceEmails.CSAServiceEmailId)
		serviceEmails.CSAServiceEmailId = ""
	}
	if companyService.NY {
		serviceEmails.NYEmailId = updateCompanyEmail(serviceEmails.NYEmailId, "NY", company.Id, companyService.NYDate, company.Email, 0, 0)
	} else if serviceEmails.NYEmailId != "" {
		db.Del("scheduled-email", serviceEmails.NYEmailId)
		serviceEmails.NYEmailId = ""
	}
	if companyService.GPS {
		serviceEmails.GPSEmailId = updateCompanyEmail(serviceEmails.GPSEmailId, "GPS", company.Id, companyService.GPSDate, company.Email, 0, 0)
	} else if serviceEmails.GPSEmailId != "" {
		db.Del("scheduled-email", serviceEmails.GPSEmailId)
		serviceEmails.GPSEmailId = ""
	}
	if companyService.Training {
		serviceEmails.TrainingEmailId = updateCompanyEmail(serviceEmails.TrainingEmailId, "Training", company.Id, companyService.TrainingDate, company.Email, 0, 0)
	} else if serviceEmails.TrainingEmailId != "" {
		db.Del("scheduled-email", serviceEmails.TrainingEmailId)
		serviceEmails.TrainingEmailId = ""
	}
	if companyService.IFTARenewal {
		serviceEmails.IFTARenewalEmailId = updateCompanyEmail(serviceEmails.IFTARenewalEmailId, "TFTA Renewal", company.Id, "11/15", company.Email, 0, 1)
	} else if serviceEmails.IFTARenewalEmailId != "" {
		db.Del("scheduled-email", serviceEmails.IFTARenewalEmailId)
		serviceEmails.IFTARenewalEmailId = ""
	}
	db.Set("company-service-emails", serviceEmails.Id, serviceEmails)
}

func updateCompanyEmail(scheduledEmailId, document, companyId, date, email string, intervalMonth, intervalYear int) string {

	//emailTime, _ := time.Parse("01/02/2006", date)

	emailTime, err := GetCompanyServiceDate(date)
	if err != nil {
		return ""
	}
	emailTS := emailTime.Unix()
	var scheduledEmail ScheduledEmail
	if !db.Get("scheduled-email", scheduledEmailId, &scheduledEmail) {
		scheduledEmailId = strconv.Itoa(int(time.Now().UnixNano()))
		scheduledEmail = ScheduledEmail{
			Id: scheduledEmailId,
			Data: map[string]Data{
				"company": Data{
					Key: "company",
					Ids: []string{companyId},
				},
			},
			Vals: map[string]interface{}{
				"document": document,
				"date":     date,
			},
			Email: mg.Email{
				To:      []string{email},
				From:    "no-reply@test.com",
				Subject: "Document Expiring",
			},
			Template: "emails/company-exp.tmpl",
		}
		if intervalMonth != 0 && intervalYear != 0 {
			scheduledEmail.Reschedule = true
			scheduledEmail.IntervalMonth = intervalMonth
			scheduledEmail.IntervalYear = intervalYear
		}
	}
	scheduledEmail.Time = emailTS
	db.Set("scheduled-email", scheduledEmailId, scheduledEmail)
	return scheduledEmailId
}

func GetCompanyServiceDate(date string) (time.Time, error) {
	now := time.Now()
	service, err := time.Parse("01/02/2006", date+"/"+strconv.Itoa(now.Year()))
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return service, err
	}
	if service.Month() < now.Month() || (service.Month() == now.Month() && service.Day() <= now.Day()) {
		service = service.AddDate(1, 0, 0)
	}
	return service, nil
}
