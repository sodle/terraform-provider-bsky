package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAccountResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testProviderPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAccountResourceConfig("test@example.com", "testpass123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("bsky_account.test", "did"),
					resource.TestCheckResourceAttr("bsky_account.test", "handle", "testusr."+pdsDomain()),
					resource.TestCheckResourceAttr("bsky_account.test", "email", "test@example.com"),
					resource.TestCheckResourceAttrSet("bsky_account.test", "password"),
				),
			},
			// ImportState testing
			{
				ResourceName: "bsky_account.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return s.RootModule().Resources["bsky_account.test"].Primary.Attributes["did"], nil
				},
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "did",
				// Password and email can't be imported
				ImportStateVerifyIgnore: []string{"password", "email"},
			},
			// Update and Read testing
			{
				Config: testAccAccountResourceConfig("updated@example.com", "newpass123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("bsky_account.test", "did"),
					resource.TestCheckResourceAttr("bsky_account.test", "handle", "testusr."+pdsDomain()),
					resource.TestCheckResourceAttr("bsky_account.test", "email", "updated@example.com"),
					resource.TestCheckResourceAttrSet("bsky_account.test", "password"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func pdsDomain() string {
	return strings.Replace(os.Getenv("BSKY_PDS_HOST"), "https://", "", 1)
}

func testAccAccountResourceConfig(email string, password string) string {
	return fmt.Sprintf(`
		resource "bsky_account" "test" {
			handle = "testusr.%[1]s"
			email    = %[2]q
			password = %[3]q
		}
	`, pdsDomain(), email, password)
}
