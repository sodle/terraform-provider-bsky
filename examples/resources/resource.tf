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

resource "bsky_list" "test-list" {
  name        = "Tf Bluesky Test List"
  purpose     = "app.bsky.graph.defs#curatelist"
  description = "Please ignore, I am testing my Tf provider."
}

resource "bsky_list_item" "comfortably-numb" {
  list_uri    = bsky_list.test-list.uri
  subject_did = "did:plc:pmyqirafcp3jqdhrl7crpq7t"
}
