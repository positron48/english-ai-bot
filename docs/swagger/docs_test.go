package swagger

import (
	"testing"

	"github.com/swaggo/swag"
)

// TestInit exercises the package init path by verifying that SwaggerInfo was
// registered with swag. The init() in docs.go runs when the package is loaded;
// this test ensures the spec is retrievable and has expected fields.
func TestInit(t *testing.T) {
	spec := swag.GetSwagger(SwaggerInfo.InstanceName())
	if spec == nil {
		t.Fatal("swag.GetSwagger(SwaggerInfo.InstanceName()) returned nil; init() may not have run or registration failed")
	}
	if SwaggerInfo.Title == "" {
		t.Error("SwaggerInfo.Title should be set")
	}
	if SwaggerInfo.Version == "" {
		t.Error("SwaggerInfo.Version should be set")
	}
}
