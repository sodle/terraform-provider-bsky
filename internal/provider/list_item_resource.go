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
	_ resource.Resource                = &listItemResource{}
	_ resource.ResourceWithConfigure   = &listItemResource{}
	_ resource.ResourceWithImportState = &listItemResource{}
)

// NewListItemResource is a helper function to simplify the provider implementation.
func NewListItemResource() resource.Resource {
	return &listItemResource{}
}

// listItemResource is the resource implementation
type listItemResource struct {
	client *xrpc.Client
}

type listItemResourceModel struct {
	Uri        types.String `tfsdk:"uri"`
	ListUri    types.String `tfsdk:"list_uri"`
	SubjectDid types.String `tfsdk:"subject_did"`
}

// Metadata returns the resource type name
func (l *listItemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list_item"
}

// Schema defines the schema for the resource.
func (r *listItemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"subject_did": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"list_uri": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"uri": schema.StringAttribute{
				Computed: true, PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (l *listItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from a plan
	var plan listItemResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Generate API request body from plan
	item := &bsky.GraphListitem{
		List:      plan.ListUri.ValueString(),
		Subject:   plan.SubjectDid.ValueString(),
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	createRecordInput := &atproto.RepoCreateRecord_Input{
		Repo:       l.client.Auth.Did,
		Collection: "app.bsky.graph.listitem",
		Record:     &util.LexiconTypeDecoder{Val: item},
	}

	// Create new list
	record, err := atproto.RepoCreateRecord(ctx, l.client, createRecordInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating list item",
			"Could not create list item, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.Uri = types.StringValue(record.Uri)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (l *listItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state listItemResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var foundItem *bsky.GraphDefs_ListItemView

	// Get refreshed list value from Bsky
	list, err := bsky.GraphGetList(ctx, l.client, "", 50, state.ListUri.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read List",
			err.Error(),
		)
		return
	}

	for _, item := range list.Items {
		if item.Uri == state.Uri.ValueString() {
			foundItem = item
			break
		}
	}

	for foundItem == nil && list.Cursor != nil {
		list, err := bsky.GraphGetList(ctx, l.client, "", 50, state.ListUri.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read List",
				err.Error(),
			)
			return
		}

		for _, item := range list.Items {
			if item.Uri == state.Uri.ValueString() {
				foundItem = item
				break
			}
		}
	}

	if foundItem == nil {
		resp.Diagnostics.AddError("List item not found", state.Uri.ValueString())
		return
	}

	state.Uri = types.StringValue(foundItem.Uri)
	state.ListUri = types.StringValue(list.List.Uri)
	state.SubjectDid = types.StringValue(foundItem.Subject.Did)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (l *listItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not implemented", "List item must be replaced, not updated")
}

// Delete deletes the resource and removes the Terraform state on success.
func (l *listItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state listItemResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing list
	uri, err := syntax.ParseATURI(state.Uri.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid list item URI",
			"Could not parse Bluesky list item URI "+state.Uri.ValueString()+": "+err.Error(),
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
			"Error deleting list item",
			"Could not delete list item, error: "+err.Error(),
		)
	}
}

// Configure adds the provider configured client to the resource.
func (l *listItemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (l *listItemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("uri"), req, resp)
}