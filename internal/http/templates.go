package http

import (
	"html/template"

	"github.com/WeKnowXD/WeKnowXD_Project/web"
)

var templates = template.Must(template.ParseFS(web.Templates,
	"templates/search.html",
	"templates/register.html",
	"templates/layout.html",
	"templates/login.html",
	"templates/weather.html",
))
