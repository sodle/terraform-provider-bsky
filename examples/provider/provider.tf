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
