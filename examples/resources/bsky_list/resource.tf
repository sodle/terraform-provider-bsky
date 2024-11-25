provider "bsky" {
  pds_host = "https://bsky.social"
  handle   = "scoott.blog"
}

resource "bsky_list" "test-list" {
  name        = "Tf Bluesky Test List"
  purpose     = "app.bsky.graph.defs#curatelist"
  description = "Please ignore, I am testing my Tf provider."
}
