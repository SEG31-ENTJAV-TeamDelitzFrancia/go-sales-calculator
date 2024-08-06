package main

import (
	"fmt"
	"html/template"
	"strings"
)

var funcMap = template.FuncMap{
	"moneyFormat": MoneyFormatter,
	"capitalise":  CapitaliseText,
}

func MoneyFormatter(money float64) string {
	return fmt.Sprintf("%12.2f", money)
}

func CapitaliseText(text string) string {
	return strings.ToUpper(text)
}
