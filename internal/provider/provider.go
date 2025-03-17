// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &bskyProvider{}

// bskyProvider defines the provider implementation.
type bskyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// bskyProviderModel maps provider schema data to a Go type.
type bskyProviderModel struct {
	PDSHost  types.String `tfsdk:"pds_host"`
	Handle   types.String `tfsdk:"handle"`
	Password types.String `tfsdk:"password"`
}

func (p *bskyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bsky"
	resp.Version = p.version
}

func (p *bskyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Bluesky PDS",
		Attributes: map[string]schema.Attribute{
			"pds_host": schema.StringAttribute{
				MarkdownDescription: "Base URL of your Personal Data Server (PDS). For most people, this is `https://bsky.social/`." +
					"\nCan also be set via the BSKY_PDS_HOST environment variable.",
				Optional: true,
			},
			"handle": schema.StringAttribute{
				MarkdownDescription: "Your Bluesky handle, without the `@`." +
					"\nCan also be set via the BSKY_HANDLE environment variable.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Your Bluesky password. Use an [app password](https://bsky.app/settings/app-passwords) for added security." +
					"\nCan also be set via the BSKY_PASSWORD environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *bskyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Bluesky client")

	// Retrieve provider data from configuration
	var config bskyProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.PDSHost.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("pds_host"),
			"Unknown Bluesky PDS host",
			"The provider cannot create the Bluesky API client as there is an unknown value for the Bluesky PDS host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BSKY_PDS_HOST environment variable.",
		)
	}
	if config.Handle.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("handle"),
			"Unknown Bluesky handle",
			"The provider cannot create the Bluesky API client as there is an unknown value for the Bluesky handle. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BSKY_HANDLE environment variable.",
		)
	}
	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Bluesky password",
			"The provider cannot create the Bluesky API client as there is an unknown value for the Bluesky password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the BSKY_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	pdsHost := os.Getenv("BSKY_PDS_HOST")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")

	if !config.PDSHost.IsNull() {
		pdsHost = config.PDSHost.ValueString()
	}

	if !config.Handle.IsNull() {
		handle = config.Handle.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if pdsHost == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("pds_host"),
			"Missing Bluesky PDS host",
			"The provider cannot create the Bluesky API client as there is a missing or empty value for the Bluesky PDS host. "+
				"Set the value in the configuration or use the BSKY_PDS_HOST environment variable."+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if handle == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("handle"),
			"Missing Bluesky handle",
			"The provider cannot create the Bluesky API client as there is a missing or empty value for the Bluesky handle. "+
				"Set the value in the configuration or use the BSKY_HANDLE environment variable."+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Bluesky password",
			"The provider cannot create the Bluesky API client as there is a missing or empty value for the Bluesky password. "+
				"Set the value in the configuration or use the BSKY_PASSWORD environment variable."+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "bluesky_pds_host", pdsHost)
	ctx = tflog.SetField(ctx, "bluesky_handle", handle)
	ctx = tflog.SetField(ctx, "bluesky_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "bluesky_password")

	tflog.Debug(ctx, "Creating Bluesky client")

	// Create a new Bluesky client with the configuration values, and log in
	client := &xrpc.Client{
		Host: pdsHost,
	}
	authInfo, err := atproto.ServerCreateSession(ctx, client, &atproto.ServerCreateSession_Input{
		Identifier: handle,
		Password:   password,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Bluesky API client",
			"An unexpected error occurred when creating the Bluesky API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"XRPC client error: "+err.Error(),
		)
	}

	client.Auth = &xrpc.AuthInfo{
		AccessJwt:  authInfo.AccessJwt,
		RefreshJwt: authInfo.RefreshJwt,
		Did:        authInfo.Did,
		Handle:     authInfo.Handle,
	}

	// Make the Bluesky client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Bluesky client", map[string]any{"success": true})
}

func (p *bskyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewListResource,
		NewListItemResource,
		NewStarterPackResource,
	}
}

func (p *bskyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewListDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &bskyProvider{
			version: version,
		}
	}
}
