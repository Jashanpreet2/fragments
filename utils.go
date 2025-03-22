package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	cognitoJwtVerify "github.com/jhosan7/cognito-jwt-verify"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func GetBody(bodyBytes []byte) map[string]interface{} {
	var jsonMap map[string]interface{}
	err := json.Unmarshal(bodyBytes, &jsonMap)
	if err != nil {
		log.Fatal(err)
	}
	return jsonMap
}

func ConvertMdToHtml(data []byte) []byte {
	extensions := parser.CommonExtensions
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(data)

	htmlFlags := html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags}
	htmlRenderer := html.NewRenderer(opts)
	return markdown.Render(doc, htmlRenderer)
}
