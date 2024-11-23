package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &listResource{}
	_ resource.ResourceWithConfigure = &listResource{}
)

// NewListResource is a helper function to simplify the provider implementation.
func NewListResource() resource.Resource {
	return &listResource{}
}

// listResource is the resource implementation
type listResource struct {
	client *xrpc.Client
}

type listResourceModel struct {
	Cid         types.String `tfsdk:"cid"`
	Uri         types.String `tfsdk:"uri"`
	Name        types.String `tfsdk:"name"`
	Purpose     types.String `tfsdk:"purpose"`
	Description types.String `tfsdk:"description"`
}

// Metadata returns the resource type name
func (l *listResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

// Schema defines the schema for the resource.
func (r *listResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cid": schema.StringAttribute{
				Computed: true,
			},
			"uri": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"purpose": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (l *listResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from a plan
	var plan listResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Generate API request body from plan
	list := &bsky.GraphList{
		Name:        plan.Name.ValueString(),
		Purpose:     plan.Purpose.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
	createRecordInput := &atproto.RepoCreateRecord_Input{
		Repo:       l.client.Auth.Did,
		Collection: "app.bsky.graph.list",
		Record:     &util.LexiconTypeDecoder{Val: list},
	}

	// Create new list
	record, err := atproto.RepoCreateRecord(ctx, l.client, createRecordInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating list",
			"Could not create list, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.Cid = types.StringValue(record.Cid)
	plan.Uri = types.StringValue(record.Uri)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (l *listResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state listResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed list value from Bsky
	list, err := bsky.GraphGetList(ctx, l.client, "", 1, state.Uri.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading list",
			"Could not read Bsky list URI "+state.Uri.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite with refreshed state
	state.Cid = types.StringValue(list.List.Cid)
	state.Name = types.StringValue(list.List.Name)
	state.Purpose = types.StringValue(*list.List.Purpose)
	state.Description = types.StringValue(*list.List.Description)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (l *listResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (l *listResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// Configure adds the provider configured client to the resource.
func (l *listResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*xrpc.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *xrpc.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	l.client = client
}
