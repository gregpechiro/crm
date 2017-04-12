package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/cagnosolutions/adb"
	"github.com/cagnosolutions/web"
	"github.com/gregpechiro/csv"
	"github.com/gregpechiro/csv/form"
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
	web.FormToStruct(&employee, r.Form, "")
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
		web.SetErrorRedirect(w, r, "/admin/customer", "Error finding customer")
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

var adminExport = web.Route{"GET", "/admin/export/:model", func(w http.ResponseWriter, r *http.Request) {
	redirect := r.Referer()
	if redirect == "" {
		redirect = "/account"
	}

	m := r.FormValue(":model")

	model, ok := CSVSINGLE[m]
	if !ok {
		web.SetErrorRedirect(w, r, redirect, "Error finding "+m+" to export")
		return
	}

	fields, err := form.GetOptions(model)
	if err != nil {
		web.SetErrorRedirect(w, r, redirect, "Error exporting "+m+"s")
		return
	}
	tc.Render(w, r, "admin-export.tmpl", web.Model{
		"fields": fields,
		"model":  m,
	})
}}

var adminExportSave = web.Route{"POST", "/admin/export/:model", func(w http.ResponseWriter, r *http.Request) {
	redirect := r.Referer()
	if redirect == "" {
		redirect = "/admin"
	}

	m := r.FormValue(":model")

	model, ok := CSVMULTI[m]
	if !ok {
		ajaxResponse(w, `{"error":true,"msg":"Error exporting `+m+`s"}`)
		return
	}

	// create pointer of model for database Unmarshel
	modelPtr := reflect.New(reflect.TypeOf(model)).Interface()

	db.All(m, modelPtr)

	// get indirect of the model pointer so the data can be read
	model = reflect.Indirect(reflect.ValueOf(modelPtr)).Interface()

	r.ParseForm()
	b, err := form.Marshal(model, r.Form)
	if err != nil {
		log.Printf("adminRoutes.go adminExportSave >> form.Marshal() >> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error exporting `+m+`s"}`)
		return
	}

	path := "export/"
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("adminRoutes.go adminExportSave >> os.MkdirAll() >> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error exporting `+m+`s"}`)
		return
	}

	path += time.Now().Format("2006-01-02") + "_" + m + "s.csv"

	if err := ioutil.WriteFile(path, b, 0666); err != nil {
		log.Printf("adminRoutes.go adminExportSave >> ioutil.WriteFile() >> %v\n", err)
		ajaxResponse(w, `{"error":true,"msg":"Error exporting `+m+`s"}`)
		return
	}

	ajaxResponse(w, `{"error":false,"path":"/admin/`+path+`"}`)
	return
}}

var adminExportDownload = web.Route{"GET", "/admin/export/:name", func(w http.ResponseWriter, r *http.Request) {
	server := http.StripPrefix("/export", http.FileServer(http.Dir("export/")))
	server.ServeHTTP(w, r)
	return
}}

var adminImportUpload = web.Route{"POST", "/admin/import/upload/:model", func(w http.ResponseWriter, r *http.Request) {
	redirect := r.Referer()
	if redirect == "" {
		redirect = "/account"
	}

	m := r.FormValue(":model")

	_, ok := CSVSINGLE[m]
	if !ok {
		web.SetErrorRedirect(w, r, redirect, "Error finding "+m+" to export")
		return
	}

	path := "import/" + m + "/"
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("main.go -> adminImportUpload -> os.MkdirAll() -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error uploading csv file")
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("main.go -> adminImporUpload -> r.FormFile() -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error uploading csv file")
		return
	}
	defer file.Close()
	f, err := os.OpenFile(path+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("main.go -> adminImportUpload -> os.OpenFile() -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error uploading csv file")
		return
	}
	defer f.Close()
	io.Copy(f, file)

	web.SetSuccessRedirect(w, r, "/admin/import/"+m+"/"+handler.Filename, "Successfully uploaded csv file")
	return
}}

var adminImport = web.Route{"GET", "/admin/import/:model/:file", func(w http.ResponseWriter, r *http.Request) {

	redirect := r.Referer()
	if redirect == "" {
		redirect = "/account"
	}

	m := r.FormValue(":model")

	model, ok := CSVSINGLE[m]
	if !ok {
		web.SetErrorRedirect(w, r, redirect, "Error finding "+m+" to import")
		return
	}

	path := "import/" + m + "/" + r.FormValue(":file")

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("main.go -> adminImport -> ioutil.ReadFile -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error finding file "+r.FormValue(":file"))
		return
	}

	dec, err := csv.NewCSVDecoder(b)
	if err != nil {
		log.Printf("main.go -> adminImport -> csv.NewCSVDecoder -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error reading file "+r.FormValue(":file"))
		return
	}

	fields, err := form.GetOptions(model)
	if err != nil {
		log.Printf("main.go -> adminImport -> form.GetOptions -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error Creating importer")
		return
	}

	header := dec.GetHeader()

	tc.Render(w, r, "admin-import.tmpl", web.Model{
		"file":   r.FormValue(":file"),
		"header": header,
		"fields": fields,
		"model":  m,
	})
	return
}}

var adminImportSave = web.Route{"POST", "/admin/import/:model/save", func(w http.ResponseWriter, r *http.Request) {

	redirect := r.Referer()
	if redirect == "" {
		redirect = "/admin"
	}

	m := r.FormValue(":model")

	model, ok := CSVMULTI[m]
	if !ok {
		web.SetErrorRedirect(w, r, redirect, "Error finding "+m+" to import")
		return
	}
	path := "import/" + m + "/" + r.FormValue("file")

	r.ParseForm()

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("main.go -> customerImportConvert -> ioutil.ReadFile -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error finding file "+r.FormValue(":file"))
		return
	}

	// create pointer of model for database Unmarshel
	modelPtr := reflect.New(reflect.TypeOf(model)).Interface()

	if err := form.Unmarshal(b, modelPtr, r.Form); err != nil {
		log.Printf("main.go -> customerImportConvert -> form.Unmarshal -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error creating "+m+"s")
		return
	}

	// get indirect of the model pointer so the data can be read
	model = reflect.Indirect(reflect.ValueOf(modelPtr)).Interface()

	/*sf, err := form.ImportableSlice(modelPtr)
	if err != nil {
		log.Printf("main.go -> customerImportConvert -> form.Importable -> %v\n", err)
		web.SetErrorRedirect(w, r, redirect, "Error creating "+m+"s")
		return
	}*/
	/*sf, k := modelPtr.([]form.Importable)
	fmt.Println(sf, k)
	if mm, ok := model.([]form.Importable); ok {
		for _, mdl := range mm {
			mdl.SetId(strconv.Itoa(int(time.Now().UnixNano())))
			db.Set(m, mdl.GetId(), mdl)
		}
		return
		web.SetErrorRedirect(w, r, redirect, "Error creating "+m+"s")
	}*/

	web.SetSuccessRedirect(w, r, redirect, "Successfully imported "+m+"s")
	return

}}

var customerImportUpload = web.Route{"POST", "/admin/customer/import", func(w http.ResponseWriter, r *http.Request) {

	path := "upload/import/customer/"
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("main.go -> customerImportUpload -> os.MkdirAll() -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error uploading csv file")
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("main.go -> customerImporUpload -> r.FormFile() -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error uploading csv file")
		return
	}
	defer file.Close()
	f, err := os.OpenFile(path+handler.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("main.go -> customerImportUpload -> os.OpenFile() -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error uploading csv file")
		return
	}
	defer f.Close()
	io.Copy(f, file)

	web.SetSuccessRedirect(w, r, "/admin/customer/import/"+handler.Filename, "Successfully uploaded csv file")
	return
}}

var customerImport = web.Route{"GET", "/admin/customer/import/:file", func(w http.ResponseWriter, r *http.Request) {
	path := "upload/import/customer/" + r.FormValue(":file")

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("main.go -> customerImport -> ioutil.ReadFile -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error finding file "+r.FormValue(":file"))
		return
	}

	dec, err := csv.NewCSVDecoder(b)
	if err != nil {
		log.Printf("main.go -> customerImport -> csv.NewCSVDecoder -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error reading file "+r.FormValue(":file"))
		return
	}
	var customer Customer

	fields, err := form.GetOptions(customer)
	if err != nil {
		log.Printf("main.go -> customerImport -> form.GetOptions -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error Creating importer")
		return
	}

	header := dec.GetHeader()
	sort.Strings(header)

	tc.Render(w, r, "admin-customer-import.tmpl", web.Model{
		"file":   r.FormValue(":file"),
		"header": header,
		"fields": fields,
	})
	return
}}

var customerImportConvert = web.Route{"POST", "/admin/customer/convert", func(w http.ResponseWriter, r *http.Request) {
	path := "upload/import/customer/" + r.FormValue("file")

	r.ParseForm()
	var customers []Customer
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("main.go -> customerImportConvert -> ioutil.ReadFile -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error finding file "+r.FormValue(":file"))
		return
	}

	if err := form.Unmarshal(b, &customers, r.Form); err != nil {
		log.Printf("main.go -> customerImportConvert -> form.Unmarshal -> %v\n", err)
		web.SetErrorRedirect(w, r, "/customer", "Error creating custoemrs")
		return
	}

	for _, customer := range customers {
		customer.Id = strconv.Itoa(int(time.Now().UnixNano()))
		db.Set("customer", customer.Id, customer)
	}

	web.SetSuccessRedirect(w, r, "/customer", "Successfully imported customers")
	return

}}
