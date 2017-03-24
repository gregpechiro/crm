package main

import "github.com/cagnosolutions/web"

var DEVELOPER = web.Auth{
	Roles:    []string{"DEVELOPER"},
	Redirect: "/login",
	Msg:      "Please Login",
}

var ADMIN = web.Auth{
	Roles:    []string{"DEVELOPER", "ADMIN"},
	Redirect: "/login",
	Msg:      "Please Login",
}

var EMPLOYEE = web.Auth{
	Roles:    []string{"DEVELOPER", "ADMIN", "EMPLOYEE"},
	Redirect: "/login",
	Msg:      "Please Login",
}

var ALL = web.Auth{
	Roles:    []string{"DEVELOPER", "ADMIN", "EMPLOYEE", "COMPANY"},
	Redirect: "/login",
	Msg:      "ERROR",
}

var COMPANY = web.Auth{
	Roles:    []string{"DEVELOPER", "COMPANY"},
	Redirect: "/customer/login",
	Msg:      "Please Login",
}
