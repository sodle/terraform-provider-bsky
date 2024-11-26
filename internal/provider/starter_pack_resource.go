package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &starterPackResource{}
	_ resource.ResourceWithConfigure   = &starterPackResource{}
	_ resource.ResourceWithImportState = &starterPackResource{}
)

// NewStarterPackResource is a helper function to simplify the provider implementation.
func NewStarterPackResource() resource.Resource {
	return &starterPackResource{}
}

// starterPackResource is the resource implementation.
type starterPackResource struct {
	client *xrpc.Client
}

type starterPackResourceModel struct {
	Uri         types.String `tfsdk:"uri"`
	ListUri     types.String `tfsdk:"list_uri"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

// Metadata returns the resource type name.
func (l *starterPackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_starter_pack"
}

// Schema defines the schema for the resource.
func (r *starterPackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Starter Packs",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The title of the Starter Pack",
				Required:            true,
			},
			"list_uri": schema.StringAttribute{
				MarkdownDescription: "The URI of the List that the Starter Pack refers too",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the Starter Pack",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "Atproto URI",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (l *starterPackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from a plan.
	var plan starterPackResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Generate API request body from plan.
	item := &bsky.GraphStarterpack{
		List:        plan.ListUri.ValueString(),
		CreatedAt:   time.Now().Format(time.RFC3339),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
	}
	createRecordInput := &atproto.RepoCreateRecord_Input{
		Repo:       l.client.Auth.Did,
		Collection: "app.bsky.graph.starterpack",
		Record:     &util.LexiconTypeDecoder{Val: item},
	}

	// Create new pack.
	record, err := atproto.RepoCreateRecord(ctx, l.client, createRecordInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating starter pack",
			"Could not create starter pack, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values.
	plan.Uri = types.StringValue(record.Uri)

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (l *starterPackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state starterPackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
	uri, err := syntax.ParseATURI(state.Uri.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid starter pack URI",
			"Could not parse Bluesky starter pack URI "+state.Uri.ValueString()+": "+err.Error(),
		)
		return
	}
	record, err := atproto.RepoGetRecord(ctx, l.client, "", uri.Collection().String(), uri.Authority().String(), uri.RecordKey().String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to retrieve starter pack",
			"Could not retrieve the current state of the starter pack "+state.Uri.ValueString()+": "+err.Error(),
		)
		return
	}
	pack, ok := record.Value.Val.(*bsky.GraphStarterpack)
	if !ok {
		resp.Diagnostics.AddError(
			"Failed to parse retrieved starter pack",
			"Could not cast the returned starter pack into the expected type",
		)
		return
	}

	state.Name = types.StringValue(pack.Name)
	state.Description = types.StringValue(*pack.Description)
	state.ListUri = types.StringValue(pack.List)

	// Set refreshed state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (l *starterPackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get current state.
	var state starterPackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
	uri, err := syntax.ParseATURI(state.Uri.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid starter pack URI",
			"Could not parse Bluesky starter pack URI "+state.Uri.ValueString()+": "+err.Error(),
		)
		return
	}
	record, err := atproto.RepoGetRecord(ctx, l.client, "", uri.Collection().String(), uri.Authority().String(), uri.RecordKey().String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to retrieve starter pack",
			"Could not retrieve the current state of the starter pack "+state.Uri.ValueString()+": "+err.Error(),
		)
		return
	}
	pack, ok := record.Value.Val.(*bsky.GraphStarterpack)
	if !ok {
		resp.Diagnostics.AddError(
			"Failed to parse retrieved starter pack",
			"Could not cast the returned starter pack into the expected type",
		)
		return
	}

	pack.Name = state.Name.ValueString()
	pack.Description = state.Description.ValueStringPointer()
	pack.List = state.ListUri.ValueString()

	// Update existing list.
	putRecordInput := &atproto.RepoPutRecord_Input{
		Collection: uri.Collection().String(),
		Repo:       uri.Authority().String(),
		Rkey:       uri.RecordKey().String(),
		SwapRecord: record.Cid,
		Record: &util.LexiconTypeDecoder{
			Val: pack,
		},
	}
	_, err = atproto.RepoPutRecord(ctx, l.client, putRecordInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update starter pack",
			"Could not update starter pack "+state.Uri.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (l *starterPackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state starterPackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing list.
	uri, err := syntax.ParseATURI(state.Uri.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid starter pack URI",
			"Could not parse Bluesky starter pack URI "+state.Uri.ValueString()+": "+err.Error(),
		)
		return
	}
	deleteRequest := &atproto.RepoDeleteRecord_Input{
		Collection: uri.Collection().String(),
		Repo:       uri.Authority().String(),
		Rkey:       uri.RecordKey().String(),
	}
	_, err = atproto.RepoDeleteRecord(ctx, l.client, deleteRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting starter pack",
			"Could not delete starter pack, error: "+err.Error(),
		)
	}
}

// Configure adds the provider configured client to the resource.
func (l *starterPackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (l *starterPackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("uri"), req, resp)
}
