package test

import (
	"os"
	"strings"
	"terraform-provider-bsky/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a new provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bsky": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func testProviderPreCheck(t *testing.T) {
	// Verify the required environment variables are set
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC is set")
	}

	v := os.Getenv("BSKY_PDS_HOST")
	if v == "" {
		t.Fatal("BSKY_PDS_HOST must be set for acceptance tests")
	}
	if !strings.HasPrefix(v, "https://") {
		t.Fatal("BSKY_PDS_HOST must start with https:// for acceptance tests")
	}
	if v := os.Getenv("BSKY_HANDLE"); v == "" {
		t.Fatal("BSKY_HANDLE must be set for acceptance tests")
	}
	if v := os.Getenv("BSKY_PASSWORD"); v == "" {
		t.Fatal("BSKY_PASSWORD must be set for acceptance tests")
	}
	if os.Getenv("BSKY_PDS_ADMIN_PASSWORD") == "" {
		t.Fatal("BSKY_PDS_ADMIN_PASSWORD must be set for acceptance tests")
	}
}
