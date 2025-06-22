package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccListItemResource(t *testing.T) {
	// Check if we should skip AppView-dependent tests
	skipAppViewTests := os.Getenv("BSKY_SKIP_APPVIEW_TESTS") != ""
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testProviderPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListItemResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("bsky_list_item.test", "uri"),
					resource.TestCheckResourceAttrSet("bsky_list_item.test", "list_uri"),
					resource.TestCheckResourceAttrSet("bsky_list_item.test", "subject_did"),
				),
			},
			// Verify the item appears in the list data source (if not skipping AppView tests)
			{
				Config: testAccListItemResourceConfig() + testAccListDataSourceConfig(),
				SkipFunc: func() (bool, error) {
					// Skip this step if we're skipping AppView tests
					return skipAppViewTests, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.bsky_list.test", "list_item_count", "1"),
					resource.TestCheckResourceAttrPair(
						"data.bsky_list.test", "items.0.did",
						"bsky_account.test", "did",
					),
				),
			},
			// ImportState testing
			{
				ResourceName: "bsky_list_item.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources["bsky_list_item.test"].Primary.Attributes["list_uri"] + "," +
						s.RootModule().Resources["bsky_list_item.test"].Primary.Attributes["uri"], nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "uri",
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccListItemResourceConfig() string {
	return fmt.Sprintf(`
		resource "bsky_list" "test" {
			name        = "Test List for Items"
			description = "A test list for the list item tests"
			purpose     = "app.bsky.graph.defs#curatelist"
		}

		resource "bsky_account" "test" {
			handle = "testusr.%[1]s"
			email = "test@example.com"
		}

		resource "bsky_list_item" "test" {
			list_uri    = bsky_list.test.uri
			subject_did = bsky_account.test.did
		}
	`, pdsDomain())
}
