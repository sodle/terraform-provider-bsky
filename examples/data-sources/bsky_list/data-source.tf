terraform {
  required_providers {
    bsky = {
      source = "sjodle/bsky"
    }
  }
}

provider "bsky" {
  pds_host = "https://bsky.social"
  handle   = "scoott.blog"
}

data "bsky_list" "bossett-science" {
  uri = "at://did:plc:jfhpnnst6flqway4eaeqzj2a/app.bsky.graph.list/3kvu7ygdzxr24"
}

output "list" {
  value = data.bsky_list.test-list
}
