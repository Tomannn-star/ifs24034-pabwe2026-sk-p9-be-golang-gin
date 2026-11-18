package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocs(t *testing.T) {
	doc1 := GetSwaggerJSON("", "")
	assert.NotEmpty(t, doc1)

	doc2 := GetSwaggerJSON("example.com:8080", "https")
	assert.Contains(t, doc2, "\"host\": \"example.com:8080\"")
	assert.Contains(t, doc2, "\"schemes\": [\"https\"]")

	assert.Equal(t, "swagger", SwaggerInfo.InstanceName())
}
