package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccListResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testProviderPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccListResourceConfig("Test List", "Test description", "app.bsky.graph.defs#curatelist"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bsky_list.test", "name", "Test List"),
					resource.TestCheckResourceAttr("bsky_list.test", "description", "Test description"),
					resource.TestCheckResourceAttr("bsky_list.test", "purpose", "app.bsky.graph.defs#curatelist"),
					resource.TestCheckResourceAttrSet("bsky_list.test", "uri"),
					resource.TestCheckResourceAttrSet("bsky_list.test", "cid"),
				),
			},
			// ImportState testing
			{
				ResourceName: "bsky_list.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources["bsky_list.test"].Primary.Attributes["uri"], nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uri",
			},
			// Update and Read testing
			{
				Config: testAccListResourceConfig("Updated List", "Updated description", "app.bsky.graph.defs#modlist"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bsky_list.test", "name", "Updated List"),
					resource.TestCheckResourceAttr("bsky_list.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("bsky_list.test", "purpose", "app.bsky.graph.defs#modlist"),
				),
			},
		},
	})
}

// Test invalid purpose.
func TestAccListResourceInvalidPurpose(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testProviderPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccListResourceConfig("Test List", "Test description", "invalid"),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

func testAccListResourceConfig(name string, description string, purpose string) string {
	return fmt.Sprintf(`
		resource "bsky_list" "test" {
			name        = %[1]q
			description = %[2]q
			purpose     = %[3]q
		}
	`, name, description, purpose)
}
