package main

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/cagnosolutions/adb"
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
	redirect := "/account"
	if employee.Home != "" {
		redirect = employee.Home
	}
	web.SetSuccessRedirect(w, r, redirect, "Welcome "+employee.FirstName)
	return
}}

var index = web.Route{"GET", "/", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/account", 303)
	return
}}

var account = web.Route{"GET", "/account", func(w http.ResponseWriter, r *http.Request) {
	employeeId := web.GetId(r)
	var employee Employee
	if !db.Get("employee", employeeId, &employee) {
		web.Logout(w)
		web.SetErrorRedirect(w, r, "/login", "Error finding your account")
		return
	}
	tc.Render(w, r, "account.tmpl", web.Model{
		"employee": employee,
	})
}}

var task = web.Route{"GET", "/task", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/task/today", 303)
	return
}}

var taskAll = web.Route{"GET", "/task/:page", func(w http.ResponseWriter, r *http.Request) {
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
	tc.Render(w, r, "task.tmpl", web.Model{
		"employee": employee,
		"tasks":    tasks,
		"page":     page,
	})
}}

var taskMarkStart = web.Route{"POST", "/task/:id/start", func(w http.ResponseWriter, r *http.Request) {
	var task Task
	db.Get("task", r.FormValue(":id"), &task)
	task.StartTime = time.Now().Unix()
	task.StopTime = 0
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, "/task/today", "Successfully Started Task")
	return
}}

var taskMarkStop = web.Route{"POST", "/task/:id/stop", func(w http.ResponseWriter, r *http.Request) {
	var task Task
	db.Get("task", r.FormValue(":id"), &task)
	task.StopTime = time.Now().Unix()
	task.TotalTime += task.StopTime - task.StartTime
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, "/task/today", "Successfully Started Task")
	return
}}

var taskMarkComplete = web.Route{"POST", "/task/:id/complete", func(w http.ResponseWriter, r *http.Request) {
	var task Task
	db.Get("task", r.FormValue(":id"), &task)
	if task.StopTime < 1 {
		task.StopTime = time.Now().Unix()
		task.TotalTime += task.StopTime - task.StartTime
	}
	task.Complete = true
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, "/task/today", "Successfully completed task")
	return
}}

var taskMarkNote = web.Route{"POST", "/task/:id/note", func(w http.ResponseWriter, r *http.Request) {
	var task Task
	db.Get("task", r.FormValue(":id"), &task)
	if r.FormValue("notes") == "" {
		web.SetErrorRedirect(w, r, "/task/today", "Error notes field was empty")
		return
	}
	task.Notes += "<li>" + r.FormValue("notes") + "</li>"
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, "/task/today", "Successfully saved notes")
	return
}}

var saveHomePage = web.Route{"POST", "/employee/:id/homepage", func(w http.ResponseWriter, r *http.Request) {
	var employee Employee
	db.Get("employee", r.FormValue(":id"), &employee)
	employee.Home = r.FormValue("url")
	db.Set("employee", employee.Id, employee)
	ajaxResponse(w, `{"error":false}`)
	return
}}

/* --- Customer Management --- */

var customerAll = web.Route{"GET", "/customer", func(w http.ResponseWriter, r *http.Request) {
	var customers []Customer
	db.All("customer", &customers)
	tc.Render(w, r, "customer-all.tmpl", web.Model{
		"customers": customers,
	})
	return
}}

var customerView = web.Route{"GET", "/customer/:id", func(w http.ResponseWriter, r *http.Request) {
	var customer Customer
	compId := r.FormValue(":id")
	if compId != "new" && !db.Get("customer", compId, &customer) {
		web.SetErrorRedirect(w, r, "/customer", "Error finding customer")
		return
	}
	var notes []Note
	var employees []Employee
	db.TestQuery("note", &notes, adb.Eq("customerId", `"`+customer.Id+`"`))

	sort.Slice(notes, func(i int, j int) bool {
		return notes[i].StartTime > notes[j].StartTime
	})

	db.All("employee", &employees)
	tc.Render(w, r, "customer.tmpl", web.Model{
		"customer":   customer,
		"notes":      notes,
		"employees":  employees,
		"quickNotes": quickNotes,
		"employeeId": web.GetId(r),
	})
	return
}}

var customerSave = web.Route{"POST", "/customer", func(w http.ResponseWriter, r *http.Request) {
	var customer Customer
	db.Get("customer", r.FormValue("id"), &customer)
	FormToStruct(&customer, r.Form, "")
	var customers []Customer
	db.TestQuery("customer", &customers, adb.Eq("email", customer.Email), adb.Ne("id", `"`+customer.Id+`"`))
	if len(customers) > 0 {
		end := "/new"
		if r.FormValue("id") != "" {
			end = "/" + r.FormValue("id")
		}
		web.SetErrorRedirect(w, r, "/customer"+end, "Error saving customer. Email is already registered")
		return
	}
	if customer.Id == "" {
		customer.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}
	if customer.SameAddress {
		customer.MailingAddress = customer.PhysicalAddress
	}

	db.Set("customer", customer.Id, customer)
	web.SetSuccessRedirect(w, r, "/customer/"+customer.Id, "Successfully saved customer")
	return
}}

var customerNoteSave = web.Route{"POST", "/customer/:id/note", func(w http.ResponseWriter, r *http.Request) {
	var note Note
	r.ParseForm()
	FormToStruct(&note, r.Form, "")
	if note.Id == "" {
		note.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}
	dt, err := time.Parse("01/02/2006 3:04 PM", r.FormValue("dateTime"))
	if err != nil {
		log.Printf("emplyeeRoutes.go >> customerSaveNotes >> time.Parse() >> %v\n", err)
	}
	note.StartTime = dt.Unix()
	note.EndTime = dt.Unix()
	note.StartTimePretty = r.FormValue("dateTime")
	note.EndTimePretty = r.FormValue("dateTime")
	db.Set("note", note.Id, note)
	web.SetSuccessRedirect(w, r, "/customer/"+r.FormValue(":id"), "Successfully saved note")
	return
}}

var customerDel = web.Route{"POST", "/customer/:id/del", func(w http.ResponseWriter, r *http.Request) {
	var customer Customer
	if !db.Get("customer", r.FormValue(":id"), &customer) {
		web.SetErrorRedirect(w, r, "/customer", "Error finding customer")
		return
	}

	// delete all notes
	var notes []Note
	db.TestQuery("note", &notes, adb.Eq("customerId", `"`+customer.Id+`"`))
	for _, note := range notes {
		db.Del("note", note.Id)
	}

	// delete customer
	db.Del("customer", customer.Id)

	web.SetSuccessRedirect(w, r, "/customer", "Successfully deleted customer")
	return
}}
