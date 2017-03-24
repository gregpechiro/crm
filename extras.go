package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cagnosolutions/web"
)

const (
	CperE = 1
	COMP  = 10
)

func defaultUsers() {

	developer := Employee{
		FirstName: "developer",
		LastName:  "developer",
	}

	developer.Id = "0"
	developer.Email = "developer@crm.com"
	developer.Password = "developer"
	developer.Active = true
	developer.Role = "DEVELOPER"

	admin := Employee{
		Id:        "1",
		FirstName: "admin",
		LastName:  "admin",
		Email:     "admin@crm.com",
		Password:  "admin",
		Active:    true,
		Role:      "ADMIN",
	}

	customer := Customer{}

	customer.Id = "0"
	customer.Name = "Test Customer"
	customer.ContactName = "Bobbi Sue"
	customer.ContactTitle = "Secretary"
	customer.ContactPhone = "717-716-6985"
	customer.PhysicalAddress.Street = "123 Main Street"
	customer.PhysicalAddress.City = "Lancaster"
	customer.PhysicalAddress.State = "PA"
	customer.PhysicalAddress.Zip = "17635"
	customer.MailingAddress.Street = "PO Box 14235"
	customer.MailingAddress.City = "Lancaster"
	customer.MailingAddress.State = "PA"
	customer.MailingAddress.Zip = "12534"
	customer.Email = "customer@test.com"

	db.Set("employee", "0", developer)
	db.Set("employee", "1", admin)
	db.Add("customer", "0", customer)

}

var makeUsers = web.Route{"GET", "/makeUsers", func(w http.ResponseWriter, r *http.Request) {
	MakeEmployees()
	MakeCustomers()
	web.SetSuccessRedirect(w, r, "/", "Successfully made users")
	return
}}

func MakeEmployees() {
	for i := 0; i < (COMP / CperE); i++ {
		id := strconv.Itoa(int(time.Now().UnixNano()))

		employee := Employee{
			FirstName: "John",
			LastName:  fmt.Sprintf("Smith the %dth", (i + 4)),
			Phone:     fmt.Sprintf("717-777-777%d", i),
		}

		employee.Id = id
		employee.Email = fmt.Sprintf("%d@cns.com", i)
		employee.Password = fmt.Sprintf("Password-%d", i)
		employee.Active = (i%2 == 0)
		employee.Role = "EMPLOYEE"

		employee.Street = fmt.Sprintf("12%d Main Street", 1)
		employee.City = fmt.Sprintf("%dville", i)
		employee.State = fmt.Sprintf("%d state", i)
		employee.Zip = fmt.Sprintf("1234%d", i)
		employee.Phone = fmt.Sprintf("717-777-777%d", i)

		db.Add("employee", id, employee)
	}
}

func MakeCustomers() [COMP]string {
	compIds := [COMP]string{}
	for i := 0; i < COMP; i++ {
		id := strconv.Itoa(int(time.Now().UnixNano()))
		compIds[i] = id

		customer := Customer{}
		customer.Id = id
		customer.Name = fmt.Sprintf("Customer %d", i)
		customer.Email = fmt.Sprintf("%d@customer%d.com", i, i)
		customer.ContactName = fmt.Sprintf("Bobbi Sue the %dth", (i + 4))
		customer.ContactTitle = fmt.Sprintf("Worker #%d", i)
		customer.ContactPhone = fmt.Sprintf("717-777-777%d", i)
		customer.ContactEmail = fmt.Sprintf("contact@customer%d.com", i)

		customer.Phone = fmt.Sprintf("717-555-555%d", i)

		customer.PhysicalAddress.Street = fmt.Sprintf("12%d Main Street", i)
		customer.PhysicalAddress.City = fmt.Sprintf("%dville", i)
		customer.PhysicalAddress.State = fmt.Sprintf("%d state", i)
		customer.PhysicalAddress.Zip = fmt.Sprintf("1234%d", i)
		if i%2 == 0 {
			customer.SameAddress = true
			customer.MailingAddress.Street = fmt.Sprintf("12%d Main Street", i)
			customer.MailingAddress.City = fmt.Sprintf("%dville", i)
			customer.MailingAddress.State = fmt.Sprintf("%d state", i)
			customer.MailingAddress.Zip = fmt.Sprintf("1234%d", i)

		} else {
			customer.SameAddress = false
			customer.MailingAddress.Street = fmt.Sprintf("12%d Main Street", i*10)
			customer.MailingAddress.City = fmt.Sprintf("%dville", i*10)
			customer.MailingAddress.State = fmt.Sprintf("%d state", i*10)
			customer.MailingAddress.Zip = fmt.Sprintf("123%d", i*10)
		}
		db.Add("customer", id, customer)
	}
	return compIds
}
