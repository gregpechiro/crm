package main

import "time"

type Address struct {
	Street string `json:"street,omitempty" csvform:""`
	City   string `json:"city,omitempty" csvform:""`
	State  string `json:"state,omitempty" csvform:""`
	Zip    string `json:"zip,omitempty" csvform:""`
}

func (a Address) AddrHTML() string {
	if a.Street == "" && a.City == "" && a.State == "" && a.Zip == "" {
		return ""
	}
	address := a.Street + "<br>" + a.City + ", "
	address += a.State + " " + a.Zip
	return address
}

type Employee struct {
	Id        string `json:"id"`
	Email     string `json:"email,omitempty" auth:"username" csv:"Email" csvform:"Email"`
	Password  string `json:"password,omitempty" auth:"password" csv:"-"`
	Active    bool   `json:"active,omitempty" auth:"active" csv:"-"`
	Role      string `json:"role,omitempty" csv:"-"`
	FirstName string `json:"firstName,omitempty" csvform:"First Name"`
	LastName  string `json:"lastName,omitempty" csvform:"Last Name"`
	Phone     string `json:"phone,omitempty"`
	Home      string `json:"home,omitempty"`
	Address   `csvform:"-"`
}

type Customer struct {
	Id              string  `json:"id" csv:"-"`
	Email           string  `json:"email,omitempty" auth:"username" csv:"email" csvform:""`
	Name            string  `json:"name,omitempty" csv:"name" csvform:"Name"`
	Phone           string  `json:"phone,omitempty" csv:"phone" csvform:""`
	ContactName     string  `json:"contactName,omitempty" csv:"contactName" csvform:"Contact Name"`
	ContactTitle    string  `json:"contactTitle,omitempty" csv:"contactTitle" csvform:"Contact Title"`
	ContactPhone    string  `json:"contactPhone,omitempty" csv:"contactPhone" csvform:"Contact Phone"`
	ContactEmail    string  `json:"contactEmail,omitempty" csv:"contactEmail" csvform:"Contact Email"`
	PhysicalAddress Address `json:"pysicalAddress,omitempty" csv:"physicalAddress" csvform:"Physical Address"`
	MailingAddress  Address `json:"mailingAddress,omitempty" csv:"mailingAddress" csvform:"Mailing Address"`
	SameAddress     bool    `json:"sameAddress"`
}

type Note struct {
	Id              string `json:"id,omitempty"`
	CustomerId      string `json:"customerId,omitempty"`
	EmployeeId      string `json:"employeeId,omitempty"`
	Communication   string `json:"communication,omitempty"`
	Purpose         string `json:"purpose,omitempty"`
	StartTime       int64  `json:"startTime,omitempty"`
	StartTimePretty string `json:"startTimePretty,omitempty"`
	EndTime         int64  `json:"endTime,omitempty"`
	EndTimePretty   string `json:"endTimePretty,omitempty"`
	Representative  string `json:"representative,omitempty"`
	CallBack        string `json:"callBack,omitempty"`
	EmailEmployee   bool   `json:"emailEmployee,omitempty"`
	Billable        bool   `json:"billable,omitempty"`
	Body            string `json:"body,omitempty"`
}

type QuickNote struct {
	Name string
	Body string
}

type Task struct {
	Id           string `json:"id"`
	EmployeeId   string `json:"employeeId,omitempty"`
	CustomerId   string `json:"customerId,omitempty"`
	CreatedTime  int64  `json:"createdTime,omitempty"`  // time.Time.Unix()
	AssignedTime int64  `json:"assignedTime,omitempty"` // time.Time.Unix()
	StartTime    int64  `json:"startTime,omitempty"`    // time.Time.Unix()
	StopTime     int64  `json:"stopTime,omitempty"`     // time.Time.Unix()
	TotalTime    int64  `json:"totalTime,ommitempty"`   // time.Time.Unix()
	Complete     bool   `json:"complete"`
	Description  string `json:"description,omitempty"`
	Notes        string `json:"notes,omitempty"`
	EmployeeName string `json:"employeeName, omitempty"`
	CustomerName string `json:"customerName, omitempty"`
}

func (t Task) GetTotalTime() int64 {
	var tt int64
	if t.StartTime > 0 {
		if t.StopTime > 0 {
			tt = t.TotalTime
		} else {
			tt = time.Now().Unix() - t.StartTime
		}
	}
	if tt < 60 {
		return 0
	}

	return tt
}

func GetTaskEmployeeView(tasks []Task) {
	for i, task := range tasks {
		var customer Customer
		db.Get("customer", task.CustomerId, &customer)
		task.CustomerName = customer.Name
		tasks[i] = task
	}
}

func GetTaskCustomerView(tasks []Task) {
	for i, task := range tasks {
		var employee Employee
		db.Get("employee", task.EmployeeId, &employee)
		task.EmployeeName = employee.FirstName + " " + employee.LastName
		tasks[i] = task
	}
}

func GetTaskAdminView(tasks []Task) {
	for i, task := range tasks {
		var customer Customer
		db.Get("customer", task.CustomerId, &customer)
		task.CustomerName = customer.Name
		var employee Employee
		db.Get("employee", task.EmployeeId, &employee)
		task.EmployeeName = employee.FirstName + " " + employee.LastName
		tasks[i] = task
	}
}
