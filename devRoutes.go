package main

import (
	"net/http"

	"github.com/cagnosolutions/web"
)

var devComments = web.Route{"GET", "/dev/comment", func(w http.ResponseWriter, r *http.Request) {
	var comments []Comment
	db.All("comment", &comments)
	tc.Render(w, r, "comment-all.tmpl", web.Model{
		"comments": comments,
	})
	return
}}

var stats = web.Route{"GET", "/dev/stats", func(w http.ResponseWriter, r *http.Request) {
	var companyServices []CompanyService
	db.All("company-service", &companyServices)
	var companyServiceEmails []CompanyServiceEmails
	db.All("company-service-emails", &companyServiceEmails)
	var companyFeatures []CompanyFeatures
	db.All("company-features", &companyFeatures)
	var notes []Note
	db.All("note", &notes)
	var notifications []Notification
	db.All("notification", &notifications)
	var documents []Document
	db.All("document", &documents)
	var drivers []Driver
	db.All("driver", &drivers)
	var vehicles []Vehicle
	db.All("vehicle", &vehicles)
	var companies []Company
	db.All("company", &companies)

	tc.Render(w, r, "stats.tmpl", web.Model{
		"companyServices":      companyServices,
		"companyServiceEmails": companyServiceEmails,
		"companyFeatures":      companyFeatures,
		"notes":                notes,
		"notifications":        notifications,
		"documents":            documents,
		"drivers":              drivers,
		"vehicles":             vehicles,
		"companies":            companies,
	})
	return
}}
