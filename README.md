# Terraform Provider for Bluesky PDS
https://registry.terraform.io/providers/sodle/bsky/latest/docs

## Getting started with the provider
Specify your PDS host url, handle, and either the password for the handle or an [app password](https://bsky.app/settings/app-passwords) for added security.
```
provider "bsky" {
  pds_host = "https://bsky.social" // or set via the BSKY_PDS_HOST env var
  handle   = "scoott.blog"         // or set via the BSKY_HANDLE   env var
  password = "<password>"          // or set via the BSKY_PASSWORD env var
}
```
## Building the provider
Install [go](https://go.dev/doc/install) and [golangci-lint](https://golangci-lint.run/welcome/install/#local-installation):
```
> choco install golang
> curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.7
```

Run the make file:
```
> make
```

## Debugging the provider
https://developer.hashicorp.com/terraform/plugin/debugging#visual-studio-code

https://developer.hashicorp.com/terraform/plugin/debugging#running-terraform-with-a-provider-in-debug-mode
