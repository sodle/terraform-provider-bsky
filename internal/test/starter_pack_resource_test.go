package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccStarterPackResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testProviderPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{ // Create and Read testing
			{
				Config: testAccStarterPackResourceConfig("Test Starter Pack", "Test description", "test1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bsky_starter_pack.test", "name", "Test Starter Pack"),
					resource.TestCheckResourceAttr("bsky_starter_pack.test", "description", "Test description"),
					resource.TestCheckResourceAttrSet("bsky_starter_pack.test", "uri"),
					resource.TestCheckResourceAttrSet("bsky_starter_pack.test", "list_uri"),
					resource.TestCheckResourceAttrPair(
						"bsky_list.test1", "uri",
						"bsky_starter_pack.test", "list_uri",
					),
				),
			},
			// ImportState testing
			{
				ResourceName: "bsky_starter_pack.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources["bsky_starter_pack.test"].Primary.Attributes["uri"], nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uri",
			},
			// Update and Read testing
			{
				Config: testAccStarterPackResourceConfig("Updated Starter Pack", "Updated description", "test2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bsky_starter_pack.test", "name", "Updated Starter Pack"),
					resource.TestCheckResourceAttr("bsky_starter_pack.test", "description", "Updated description"),
					resource.TestCheckResourceAttrSet("bsky_starter_pack.test", "uri"),
					resource.TestCheckResourceAttrSet("bsky_starter_pack.test", "list_uri"),
					resource.TestCheckResourceAttrPair(
						"bsky_list.test2", "uri",
						"bsky_starter_pack.test", "list_uri",
					),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccStarterPackResourceConfig(name string, description string, listName string) string {
	return fmt.Sprintf(`
resource "bsky_list" "test1" {
	name        = "test list for starter pack"
	description = "A list for reference in other resources"
	purpose     = "app.bsky.graph.defs#curatelist"
}

resource "bsky_list" "test2" {
	name        = "test list for starter pack update"
	description = "A list for reference in other resources"
	purpose     = "app.bsky.graph.defs#curatelist"
}

resource "bsky_starter_pack" "test" {
	name        = %[1]q
	description = %[2]q
	list_uri    = bsky_list.%[3]s.uri
}
`, name, description, listName)
}
