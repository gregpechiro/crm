package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cagnosolutions/adb"
	"github.com/cagnosolutions/web"
)

/* --- Employee Management --- */

var employeeAll = web.Route{"GET", "/admin/employee", func(w http.ResponseWriter, r *http.Request) {
	var employees []Employee
	// get all "employees" except the default logins
	db.TestQuery("employee", &employees, adb.Gt("id", `"1"`))
	tc.Render(w, r, "admin-employee-all.tmpl", web.Model{
		"employees": employees,
	})
	return
}}

var employeeView = web.Route{"GET", "/admin/employee/:id", func(w http.ResponseWriter, r *http.Request) {
	var employee Employee
	employeeId := r.FormValue(":id")
	if employeeId != "new" && !db.Get("employee", employeeId, &employee) {
		web.SetErrorRedirect(w, r, "/admin/employee", "Error finding employee")
		return
	}
	tc.Render(w, r, "admin-employee.tmpl", web.Model{
		"employee": employee,
	})
	return
}}

var employeeSave = web.Route{"POST", "/admin/employee", func(w http.ResponseWriter, r *http.Request) {
	empId := r.FormValue("id")
	var employee Employee
	db.Get("employee", empId, &employee)
	FormToStruct(&employee, r.Form, "")
	if employee.Id == "" && empId == "" {
		employee.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}

	var employees []Employee
	db.TestQuery("employee", &employees, adb.Eq("email", employee.Email), adb.Ne("id", `"`+employee.Id+`"`))
	if len(employees) > 0 {
		web.SetErrorRedirect(w, r, "/admin/employee/"+employee.Id, "Error saving employee. Email is already registered")
		return
	}
	db.Set("employee", employee.Id, employee)
	web.SetSuccessRedirect(w, r, "/admin/employee/"+employee.Id, "Successfully saved employee")
	return
}}

var adminEmployeeTask = web.Route{"GET", "/admin/employee/:id/task", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/employee/"+r.FormValue(":id")+"/task/today", 303)
	return
}}

var adminEmployeeTaskAll = web.Route{"GET", "/admin/employee/:id/task/:page", func(w http.ResponseWriter, r *http.Request) {
	employeeId := r.FormValue(":id")
	var employee Employee
	if !db.Get("employee", employeeId, &employee) {
		web.SetErrorRedirect(w, r, "/admin/employee", "Error finding employee")
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
	var customers []Customer
	db.All("customer", &customers)
	tc.Render(w, r, "admin-employee-task.tmpl", web.Model{
		"employee": employee,
		"customer": customers,
		"tasks":    tasks,
		"page":     page,
	})
}}

var employeeDel = web.Route{"POST", "/admin/employee/:id", func(w http.ResponseWriter, r *http.Request) {
	empId := r.FormValue(":id")
	db.Del("employee", empId)
	web.SetSuccessRedirect(w, r, "/admin/employee", "Successfully deleted employee")
	return
}}

var adminTask = web.Route{"GET", "/admin/task", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/task/all", 303)
	return
}}

var adminTaskAll = web.Route{"GET", "/admin/task/:page", func(w http.ResponseWriter, r *http.Request) {
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
	case "today":
		beg, end := Today()
		db.TestQuery("task", &tasks, adb.Gt("assignedTime", strconv.Itoa(int(beg))), adb.Lt("assignedTime", strconv.Itoa(int(end))))
	case "overdue":
		beg, _ := Today()
		db.TestQuery("task", &tasks, adb.Lt("assignedTime", strconv.Itoa(int(beg))), adb.Eq("complete", "false"))
	case "incomplete":
		db.TestQuery("task", &tasks, adb.Eq("complete", "false"))
	case "complete":
		db.TestQuery("task", &tasks, adb.Eq("complete", "true"))
	default:
		page = "all"
		db.All("task", &tasks)
	}

	GetTaskAdminView(tasks)
	var employees []Employee
	db.All("employee", &employees)
	var customers []Customer
	db.All("customer", &customers)

	tc.Render(w, r, "admin-task.tmpl", web.Model{
		"customers": customers,
		"employee":  employee,
		"employees": employees,
		"tasks":     tasks,
		"page":      page,
	})
}}

var adminTasksave = web.Route{"POST", "/admin/task", func(w http.ResponseWriter, r *http.Request) {
	taskId := r.FormValue("id")
	var task Task
	if taskId != "" {
		db.Get("task", taskId, &task)
	} else {
		task.Id = strconv.Itoa(int(time.Now().UnixNano()))
	}
	redirect := r.FormValue("redirect")
	if redirect == "" {
		redirect = "/admin/task/all"
	}
	loc, _ := time.LoadLocation("Local")
	t, err := time.ParseInLocation("1/2/2006", r.FormValue("assignedDate"), loc)
	if err != nil {
		log.Printf("adminRoutes.go >> adminTaskSave >> ParseInLocation >> %v\n\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error saving task.")
		return
	}

	task.Description = r.FormValue("description")
	task.EmployeeId = r.FormValue("employeeId")
	task.CustomerId = r.FormValue("customerId")
	task.AssignedTime = t.Unix()
	task.CreatedTime = time.Now().Unix()
	db.Set("task", task.Id, task)
	web.SetSuccessRedirect(w, r, redirect, "Successfully saved task")
	return

}}

var adminCustomerTask = web.Route{"GET", "/admin/customer/:id/task", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin/customer/"+r.FormValue(":id")+"/task/today", 303)
	return
}}

var adminCustomerTaskAll = web.Route{"GET", "/admin/customer/:id/task/:page", func(w http.ResponseWriter, r *http.Request) {
	customerId := r.FormValue(":id")
	var customer Customer
	if !db.Get("customer", customerId, &customer) {
		web.SetErrorRedirect(w, r, "/cns/copany", "Error finding customer")
		return
	}
	var tasks []Task
	page := r.FormValue(":page")

	switch page {
	case "overdue":
		beg, _ := Today()
		db.TestQuery("task", &tasks, adb.Eq("customerId", `"`+customer.Id+`"`), adb.Lt("assignedTime", strconv.Itoa(int(beg))), adb.Eq("complete", "false"))
	case "incomplete":
		db.TestQuery("task", &tasks, adb.Eq("customerId", `"`+customer.Id+`"`), adb.Eq("complete", "false"))
	case "complete":
		db.TestQuery("task", &tasks, adb.Eq("customerId", `"`+customer.Id+`"`), adb.Eq("complete", "true"))
	default:
		page = "today"
		beg, end := Today()
		db.TestQuery("task", &tasks, adb.Eq("customerId", `"`+customer.Id+`"`), adb.Gt("assignedTime", strconv.Itoa(int(beg))), adb.Lt("assignedTime", strconv.Itoa(int(end))))
	}

	GetTaskCustomerView(tasks)
	var employees []Employee
	db.All("employee", &employees)
	tc.Render(w, r, "admin-customer-task.tmpl", web.Model{
		"customer":  customer,
		"employees": employees,
		"tasks":     tasks,
		"page":      page,
	})
}}
