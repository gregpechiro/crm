package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cagnosolutions/adb"
	"github.com/cagnosolutions/web"
)

var customerLogin = web.Route{"GET", "/customer/login", func(w http.ResponseWriter, r *http.Request) {
	tc.Render(w, r, "customer-login.tmpl", web.Model{})
	return
}}

var customerLoginPost = web.Route{"POST", "/customer/login", func(w http.ResponseWriter, r *http.Request) {
	email, pass := r.FormValue("email"), r.FormValue("password")
	var company Company
	if !db.Auth("company", email, pass, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Incorrect username or password")
		return
	}
	sess := web.Login(w, r, company.Role)
	sess.PutId(w, company.Id)
	sess["email"] = company.Email
	sess["collapse"] = false
	web.PutMultiSess(w, r, sess)
	web.SetSuccessRedirect(w, r, "/customer", "Welcome")
	return

}}

var customerLogout = web.Route{"GET", "/customer/logout", func(w http.ResponseWriter, r *http.Request) {
	web.Logout(w)
	http.Redirect(w, r, "/customer/login", 303)
	return
}}

var customerHome = web.Route{"GET", "/customer", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var notifications []Notification
	db.TestQuery("notification", &notifications, adb.Eq("type", "COMPANY"), adb.Eq("modelId", `"`+company.Id+`"`))
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+company.Id+`"`))
	tc.Render(w, r, "customer-home.tmpl", web.Model{
		"company":         company,
		"notifications":   notifications,
		"companyFeatures": companyFeatures,
	})
	return
}}

var customerInfo = web.Route{"GET", "/customer/info", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+company.Id+`"`))
	wm := web.Model{
		"company":         company,
		"companyFeatures": companyFeatures,
	}
	passwordStat := r.FormValue("password")
	if !company.PasswordCheck() {
		passwordStat = "force"
		wm["alertWarning"] = "Please change your password before continuing"
	}
	wm["passwordStat"] = passwordStat
	tc.Render(w, r, "customer-info.tmpl", wm)
}}

var customerPasswordSave = web.Route{"POST", "/customer/password", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	password := r.FormValue("password")
	if password == "" {
		web.SetErrorRedirect(w, r, "/customer/info?password=change", "Password cannot be blank")
		return
	}
	if password == company.Email {
		web.SetErrorRedirect(w, r, "/customer/info?password=change", "Password must be different than email address")
		return
	}
	company.Password = password
	db.Set("company", company.Id, company)
	web.SetSuccessRedirect(w, r, "/customer/info", "Successfully changed password")
	return
}}

var customerVehicle = web.Route{"GET", "/customer/vehicle", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var vehicles []Vehicle
	db.TestQuery("vehicle", &vehicles, adb.Eq("companyId", `"`+id+`"`))
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))
	tc.Render(w, r, "customer-vehicle-all.tmpl", web.Model{
		"company":         company,
		"vehicles":        vehicles,
		"companyFeatures": companyFeatures,
	})
	return
}}

var customerDriver = web.Route{"GET", "/customer/driver", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var drivers []Driver
	db.TestQuery("driver", &drivers, adb.Eq("companyId", `"`+company.Id+`"`))
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+company.Id+`"`))
	tc.Render(w, r, "customer-driver-all.tmpl", web.Model{
		"company":         company,
		"drivers":         drivers,
		"companyFeatures": companyFeatures,
	})
	return
}}

var customerForm = web.Route{"GET", "/customer/form", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var docs []Document
	db.TestQuery("document", &docs, adb.Eq("companyId", `"`+company.Id+`"`), adb.Eq("stateForm", "true"))
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+company.Id+`"`))
	tc.Render(w, r, "customer-form.tmpl", web.Model{
		"company":         company,
		"docs":            docs,
		"companyFeatures": companyFeatures,
	})
	return
}}

var customerFile = web.Route{"GET", "/customer/file", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
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
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+company.Id+`"`))

	tc.Render(w, r, "customer-file.tmpl", web.Model{
		"company":         company,
		"companyFeatures": companyFeatures,
	})
}}

var customerVehicleView = web.Route{"GET", "/customer/vehicle/:id", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var vehicle Vehicle
	if !db.Get("vehicle", r.FormValue(":id"), &vehicle) || vehicle.CompanyId != id {
		web.SetErrorRedirect(w, r, "/customer/vehicle", "Error finding vehicle")
		return
	}
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))
	tc.Render(w, r, "customer-vehicle.tmpl", web.Model{
		"vehicle":         vehicle,
		"companyFeatures": companyFeatures,
	})
	return
}}

var customerVehicleFile = web.Route{"GET", "/customer/vehicle/:id/file", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var vehicle Vehicle
	if !db.Get("vehicle", r.FormValue(":id"), &vehicle) || vehicle.CompanyId != id {
		web.SetErrorRedirect(w, r, "/customer/vehicle", "Error finding vehicle")
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
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))
	tc.Render(w, r, "customer-vehicle-file.tmpl", web.Model{
		"vehicle":         vehicle,
		"files":           files,
		"companyFeatures": companyFeatures,
	})
}}

var customerDriverView = web.Route{"GET", "/customer/driver/:id", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var driver Driver
	if !db.Get("driver", r.FormValue(":id"), &driver) || driver.CompanyId != id {
		web.SetErrorRedirect(w, r, "/customer/driver", "error finding driver")
		return
	}
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))
	tc.Render(w, r, "customer-driver.tmpl", web.Model{
		"driver":          driver,
		"companyFeatures": companyFeatures,
	})
	return
}}

var customerDriverForm = web.Route{"GET", "/customer/driver/:id/form", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var driver Driver
	if !db.Get("driver", r.FormValue(":id"), &driver) || driver.CompanyId != id {
		web.SetErrorRedirect(w, r, "/customer/driver", "error finding driver")
		return
	}
	var docs []Document
	db.TestQuery("document", &docs, adb.Eq("driverId", `"`+driver.Id+`"`), adb.Eq("companyId", `"`+driver.CompanyId+`"`))
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))

	tc.Render(w, r, "customer-driver-form.tmpl", web.Model{
		"driver":          driver,
		"docs":            docs,
		"companyFeatures": companyFeatures,
	})
}}

var customerDriverFile = web.Route{"GET", "/customer/driver/:id/file", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var driver Driver
	if !db.Get("driver", r.FormValue(":id"), &driver) || driver.CompanyId != id {
		web.SetErrorRedirect(w, r, "/customer/driver", "error finding driver")
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
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))
	tc.Render(w, r, "customer-driver-file.tmpl", web.Model{
		"driver":          driver,
		"files":           files,
		"companyFeatures": companyFeatures,
	})
}}

var customerViolation = web.Route{"GET", "/customer/violation", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))
	tc.Render(w, r, "customer-violation.tmpl", web.Model{
		"company":         company,
		"companyFeatures": companyFeatures,
		"violations":      GetCustomerViolations(id),
	})
}}

var customerSafer = web.Route{"GET", "/customer/safer", func(w http.ResponseWriter, r *http.Request) {
	id := web.GetId(r)
	var company Company
	if !db.Get("company", id, &company) {
		web.SetErrorRedirect(w, r, "/customer/login", "Error finding your account")
		web.Logout(w)
		return
	}
	if !company.PasswordCheck() {
		web.SetFlashRedirect(w, r, "/customer/info?password=force", "alertWarning", "Please change your password before continuing")
		return
	}
	var companyFeatures CompanyFeatures
	db.TestQueryOne("company-features", &companyFeatures, adb.Eq("companyId", `"`+id+`"`))
	tc.Render(w, r, "customer-safer.tmpl", web.Model{
		"company":         company,
		"companyFeatures": companyFeatures,
		"safer":           GetCustomerSafer(id),
	})
}}

/*var customerViolationRest = web.Route{"GET", "/customer/:id/violation", func(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue(":id")
	var company Company
	db.Get("company", id, &company)
	var violations ViolationCache
	var resp *http.Response
	var err error
	var b []byte
	restResp := make(map[string]interface{})
	restResp["cache"] = true

	if true || !db.Get("violation-cache", id, &violations) || violations.NeedsUpdate() {

		resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/UnsafeDriving.aspx")
		if err != nil {
			goto mark
		}
		defer resp.Body.Close()
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			goto mark
		}
		unsafeDriving := base64.StdEncoding.EncodeToString(b)

		resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/DrugsAlcohol.aspx")
		if err != nil {
			goto mark
		}
		defer resp.Body.Close()
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			goto mark
		}
		controlledSubstances := base64.StdEncoding.EncodeToString(b)

		resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/DriverFitness.aspx")
		if err != nil {
			goto mark
		}
		defer resp.Body.Close()
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			goto mark
		}
		driverFitness := base64.StdEncoding.EncodeToString(b)

		resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/HOSCompliance.aspx")
		if err != nil {
			goto mark
		}
		defer resp.Body.Close()
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			goto mark
		}
		hosCompliance := base64.StdEncoding.EncodeToString(b)

		resp, err = http.Get("https://ai.fmcsa.dot.gov/SMS/Carrier/" + company.DOTNum + "/BASIC/VehicleMaint.aspx")
		if err != nil {
			goto mark
		}
		defer resp.Body.Close()
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			goto mark
		}
		vehicleMaintenance := base64.StdEncoding.EncodeToString(b)

		restResp["cache"] = false
		violations.LastUpdate = time.Now().Format("01/02/2006")
		violations.UnsafeDrivingBody = unsafeDriving
		violations.ControlledSubstancesBody = controlledSubstances
		violations.DriverFitnessBody = driverFitness
		violations.HOSComplianceBody = hosCompliance
		violations.VehicleMaintenanceBody = vehicleMaintenance
	}

mark:

	restResp["violations"] = violations
	b, err = json.Marshal(restResp)
	if err != nil {
		ajaxResponse(w, `{"error":true}`)
		return
	}
	ajaxResponse(w, string(b))

}}*/

var customerViolationRestSave = web.Route{"POST", "/customer/:id/violation/rest", func(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue(":id")
	r.ParseForm()
	var violationCache ViolationCache
	FormToStruct(&violationCache, r.Form, "")
	db.Set("violation-cache", id, violationCache)
}}

var customerSaferRestSave = web.Route{"POST", "/customer/:id/safer/rest", func(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue(":id")
	r.ParseForm()
	var saferCache SaferCache
	FormToStruct(&saferCache, r.Form, "")
	db.Set("safer-cache", id, saferCache)
}}
