package main

import (
	"html/template"
	"net/http"
	"strconv"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	// create view
	t, err := template.New("index.html").ParseFiles(filepath + "/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// serve the template
	if err = t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func CalculatePartialHandler(w http.ResponseWriter, r *http.Request) {
	var e Employee
	var err error

	// must run ParseForm before accessing formdata
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// set form data
	// TODO: find a better way to do this
	e.EmployeeID = r.FormValue("employeeID")
	e.EmployeeName = r.FormValue("employeeName")
	e.SalesCode = r.FormValue("salesCode")
	e.SalesAmount, err = strconv.ParseFloat(r.FormValue("salesAmount"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// internal logic, supposedly
	if err := e.ComputeTakeHomePay(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// log.Println(e)

	// create partial
	t, err := template.New("display-payroll.html").Funcs(funcMap).ParseFiles(filepath + "/partials/display-payroll.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// run partial as response
	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
