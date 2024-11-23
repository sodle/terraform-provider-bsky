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
  name        = "Tf Bluesky Test List Updated"
  purpose     = "app.bsky.graph.defs#modlist"
  description = "Please ignore, I am testing my Tf provider. Updated description."
}
