package main

import "errors"

type Employee struct {
	EmployeeID        string // input
	EmployeeName      string
	SalesCode         string
	SalesAmount       float64
	TakeHomePay       float64 // calculated
	GrossEarnedAmount float64
	SalesCommission   float64
}

func (e *Employee) computeSalesCodeAlpha() {
	e.GrossEarnedAmount = SalesCodeAlphaAdditional + (SalesCodeAlphaRate * e.SalesAmount)
}
func (e *Employee) computeSalesCodeBravo() {
	e.GrossEarnedAmount = SalesCodeBravoAdditional + (SalesCodeBravoRate * e.SalesAmount)
}
func (e *Employee) computeSalesCodeCharlie() {
	e.GrossEarnedAmount = SalesCodeCharlieAdditional + (SalesCodeCharlieRate * e.SalesAmount)
}

func (e *Employee) computeSalesAdditionalCommission() {
	if e.SalesAmount > MinimumCommission {
		e.SalesCommission = AdditionalSalesCommission * e.SalesAmount
	}
}

func (e *Employee) ComputeTakeHomePay() error {

	e.computeSalesAdditionalCommission()

	switch e.SalesCode {
	case "a":
		e.computeSalesCodeAlpha()
	case "b":
		e.computeSalesCodeBravo()
	case "c":
		e.computeSalesCodeCharlie()
	default:
		return errors.New("invalid Sales Code")
	}

	e.TakeHomePay = e.GrossEarnedAmount + e.SalesCommission
	return nil
}
