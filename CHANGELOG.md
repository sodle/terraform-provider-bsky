## 1.4.0

FEATURES: 

- Added Terraform acceptance tests

BUG FIXES:

- Nil reference for bsky_list data source when optional attributes are not set.
- bsky_list.purpose needed a validator
- During provider initialization when the Bluesky API client cannot be created the provider should exit.
- In bsky_starter_pack Update the updated name, description and list were being taken from the state instead of plan.
- Breaking Change: list_item import needs to specify the list_uri + list_item_uri as there is no API to get the list_uri from the list_item_uri. This technically isn't a breaking change as import wasn't working before.

REFACTORS:

- Needed to switch to RepoGetRecord for list. This is because the GraphGetList uses an AppView and we do not have an AppView configured in the GitHub action tests. Most of the properties of the list are available from the record - the list item count and items are not so they are skipped in the acceptance tests. These tests still run when using a real PDS for testing.
- Migrated to golangci-lint v2


## 1.2.0

FEATURES:

- New resource: `bsky_account` (contribution by [@alangoldman](https://github.com/alangoldman))

## 1.1.0

FEATURES:

- New resource: `bsky_starter_pack`

## 1.0.0

FEATURES:

- Initial release
- New resources: `bsky_list`, `bsky_list_item`
- New data source: `bsky_list`
