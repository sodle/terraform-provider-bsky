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

data "bsky_list" "test-list" {
  uri = "at://did:plc:pmyqirafcp3jqdhrl7crpq7t/app.bsky.graph.list/3lam62tvlqz2l"
}

output "list" {
  value = data.bsky_list.test-list
}
