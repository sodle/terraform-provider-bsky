provider "bsky" {
  pds_host = "https://bsky.social"
  handle   = "scoott.blog"
}

resource "bsky_list" "test-list" {
  name        = "Tf Bluesky Test List"
  purpose     = "app.bsky.graph.defs#curatelist"
  description = "Please ignore, I am testing my Tf provider."
}

resource "bsky_starter_pack" "test-pack" {
  name        = "Tf Bluesky Starter Pack"
  description = "Test, please ignore"
  list_uri    = bsky_list.test-list.uri
}
