package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccListDataSource(t *testing.T) {
	// Check if we should skip AppView-dependent tests
	skipAppViewTests := os.Getenv("BSKY_SKIP_APPVIEW_TESTS") != ""
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testProviderPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories, Steps: []resource.TestStep{
			// Create the list first
			{
				Config: testAccListBaseConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("bsky_list.test", "uri"),
				),
			},
			// Now read it with the data source and verify basic attributes
			{
				Config: testAccListBaseConfig() + testAccListDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check base attributes
					resource.TestCheckResourceAttr("data.bsky_list.test", "name", "Test List for Data Source"),
					resource.TestCheckResourceAttr("data.bsky_list.test", "description", "A test list for the data source tests"),
					resource.TestCheckResourceAttr("data.bsky_list.test", "purpose", "app.bsky.graph.defs#curatelist"),
					resource.TestCheckResourceAttrSet("data.bsky_list.test", "uri"),
					resource.TestCheckResourceAttrSet("data.bsky_list.test", "cid"),
					// Check optional attributes
					resource.TestCheckResourceAttr("data.bsky_list.test", "avatar", ""),
				),
			},
			// Check AppView-dependent attributes (empty list)
			{
				Config: testAccListBaseConfig() + testAccListDataSourceConfig(),
				SkipFunc: func() (bool, error) {
					// Skip this step if we're skipping AppView tests
					return skipAppViewTests, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check AppView-dependent attributes
					resource.TestCheckResourceAttr("data.bsky_list.test", "list_item_count", "0"),
					// Check empty items array
					resource.TestCheckResourceAttr("data.bsky_list.test", "items.#", "0"),
				),
			},
			// Add an item to the list in a separate step (adding the item to the list is asynchronous)
			{
				Config: testAccListBaseConfig() + testAccListDataSourceWithItemConfig(),
			},
			// Now read it again and verify basic attributes still match
			{
				Config: testAccListBaseConfig() + testAccListDataSourceWithItemConfig() + testAccListDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check base attributes still match
					resource.TestCheckResourceAttr("data.bsky_list.test", "name", "Test List for Data Source"),
					resource.TestCheckResourceAttr("data.bsky_list.test", "description", "A test list for the data source tests"),
					resource.TestCheckResourceAttr("data.bsky_list.test", "purpose", "app.bsky.graph.defs#curatelist"),
					resource.TestCheckResourceAttrSet("data.bsky_list.test", "uri"),
					resource.TestCheckResourceAttrSet("data.bsky_list.test", "cid"),
				),
			},
			// Check AppView-dependent attributes (with list item)
			{
				Config: testAccListBaseConfig() + testAccListDataSourceWithItemConfig() + testAccListDataSourceConfig(),
				SkipFunc: func() (bool, error) {
					// Skip this step if we're skipping AppView tests
					return skipAppViewTests, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check the item count and list items
					resource.TestCheckResourceAttr("data.bsky_list.test", "list_item_count", "1"),
					resource.TestCheckResourceAttr("data.bsky_list.test", "items.#", "1"),
					resource.TestCheckResourceAttrSet("data.bsky_list.test", "items.0.did"),
					resource.TestCheckResourceAttrSet("data.bsky_list.test", "items.0.uri"),
				),
			},
		},
	})
}

func testAccListBaseConfig() string {
	return `
		resource "bsky_list" "test" {
			name        = "Test List for Data Source"
			description = "A test list for the data source tests"
			purpose     = "app.bsky.graph.defs#curatelist"
		}
	`
}

func testAccListDataSourceConfig() string {
	return `
	data "bsky_list" "test" {
		uri = bsky_list.test.uri
	}`
}

func testAccListDataSourceWithItemConfig() string {
	return fmt.Sprintf(`
		resource "bsky_account" "test" {
			handle = "testusr.%[1]s"
			email = "test@example.com"
		}

		resource "bsky_list_item" "test" {
			list_uri = bsky_list.test.uri
			subject_did = bsky_account.test.did
		}
	`, pdsDomain())
}
