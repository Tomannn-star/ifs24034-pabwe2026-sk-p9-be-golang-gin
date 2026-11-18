package docs

import (
	_ "embed"
	"strings"

	"github.com/swaggo/swag"
)

//go:embed swagger.json
var SwaggerJSON string

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8081",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Delcom Blog API",
	Description:      "Dokumentasi REST API Delcom Blog Multi-Tier & AI Generator dengan Integrasi Search Engine",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  SwaggerJSON,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

func GetSwaggerJSON(host, scheme string) string {
	doc := SwaggerJSON
	if host != "" {
		doc = strings.Replace(doc, "\"host\": \"localhost:8081\"", "\"host\": \""+host+"\"", 1)
	}
	if scheme != "" {
		doc = strings.Replace(doc, "\"schemes\": [\n    \"http\"\n  ]", "\"schemes\": [\""+scheme+"\"]", 1)
	}
	return doc
}
