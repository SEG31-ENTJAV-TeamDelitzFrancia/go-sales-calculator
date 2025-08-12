package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const otelName = "go-sales-calculator"

var (
	tracer        = otel.Tracer(otelName)
	meter         = otel.Meter(otelName)
	logger        = otelslog.NewLogger(otelName)
	calculatedCtr metric.Int64Counter
)

func init() {
	var err error
	calculatedCtr, err = meter.Int64Counter("requests",
		metric.WithDescription("The number of requests made for each calculation"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		panic(err)
	}
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	// start otelling
	ctx, span := tracer.Start(r.Context(), "root")
	defer span.End()

	// create view
	t, err := template.New("index.html").ParseFS(htmlFS, "index.html")
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
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
	// start otelling
	ctx, span := tracer.Start(r.Context(), "calculate")
	defer span.End()

	var e Employee
	var err error

	// must run ParseForm before accessing formdata
	if err := r.ParseForm(); err != nil {
		logger.ErrorContext(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// set form data
	// TODO: find a better way to do this
	e.EmployeeID = r.FormValue("employeeID")
	e.EmployeeName = r.FormValue("employeeName")
	e.SalesCode = r.FormValue("salesCode")
	e.SalesAmount, err = strconv.ParseFloat(r.FormValue("salesAmount"), 64)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// internal logic, supposedly
	if err := e.ComputeTakeHomePay(); err != nil {
		logger.ErrorContext(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	logger.InfoContext(ctx, "Successfully Calculated Employee Pay", "employee", e)

	// marshal calculated employee as otel attribute
	jsonString, err := json.Marshal(e)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	employeeAttr := attribute.String("calculate.employee", string(jsonString))
	span.SetAttributes(employeeAttr)
	calculatedCtr.Add(ctx, 1, metric.WithAttributes(employeeAttr))

	// create partial
	t, err := template.New("display-payroll.html").Funcs(funcMap).ParseFS(htmlFS, "partials/display-payroll.html")
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// run partial as response
	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, e)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
