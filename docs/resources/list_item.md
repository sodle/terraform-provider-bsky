---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "bsky_list_item Resource - bsky"
subcategory: ""
description: |-
  Manage users' membership on Bluesky lists
---

# bsky_list_item (Resource)

Manage users' membership on Bluesky lists

## Example Usage

```terraform
provider "bsky" {
  pds_host = "https://bsky.social"
  handle   = "scoott.blog"
}

resource "bsky_list" "test-list" {
  name        = "Tf Bluesky Test List"
  purpose     = "app.bsky.graph.defs#curatelist"
  description = "Please ignore, I am testing my Tf provider."
}

resource "bsky_list_item" "scoott" {
  list_uri    = bsky_list.test-list.uri
  subject_did = "did:plc:7kkf4hujjl6wll6pewqahaex"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `list_uri` (String) The URI of the list
- `subject_did` (String) The DID of the user to add to the list

### Read-Only

- `uri` (String) Atproto URI

## Import

Import is supported using the following syntax:

```shell
# List item can be imported using the format ListURI,ListItemURI.
terraform import bsky_list_item.scoott "at://did:plc:7kkf4hujjl6wll6pewqahaex/app.bsky.graph.list/3lbo5zov45j2q,at://did:plc:7kkf4hujjl6wll6pewqahaex/app.bsky.graph.listitem/3lbqcyq3uzo2u"
```
