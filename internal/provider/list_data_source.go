package provider

import (
	"context"
	"fmt"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &listDataSource{}
	_ datasource.DataSourceWithConfigure = &listDataSource{}
)

// NewListDataSource is a helper function to simplify the provider implementation.
func NewListDataSource() datasource.DataSource {
	return &listDataSource{}
}

// listDataSource is the data source implementation.
type listDataSource struct {
	client *xrpc.Client
}

// listItemModel represents an item in a list.
type listItemModel struct {
	Did types.String `tfsdk:"did"`
	Uri types.String `tfsdk:"uri"`
}

// listDataSourceModel maps the data source schema data.
type listDataSourceModel struct {
	Avatar        types.String `tfsdk:"avatar"`
	Cid           types.String `tfsdk:"cid"`
	Description   types.String `tfsdk:"description"`
	ListItemCount types.Int64  `tfsdk:"list_item_count"`
	Name          types.String `tfsdk:"name"`
	Purpose       types.String `tfsdk:"purpose"`
	Uri           types.String `tfsdk:"uri"`

	Items []listItemModel `tfsdk:"items"`
}

// Metadata returns the data source type name.
func (d *listDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

// Schema defines the schema for the data source.
func (d *listDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"avatar": schema.StringAttribute{
				Computed: true,
			},
			"cid": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"list_item_count": schema.Int64Attribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"purpose": schema.StringAttribute{
				Computed: true,
			},
			"uri": schema.StringAttribute{
				Required: true,
			},

			"items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"did": schema.StringAttribute{
							Computed: true,
						},
						"uri": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *listDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	uri := data.Uri.ValueString()

	list, err := bsky.GraphGetList(ctx, d.client, "", 50, uri)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read List",
			err.Error(),
		)
		return
	}

	data.Avatar = types.StringValue(*list.List.Avatar)
	data.Cid = types.StringValue(list.List.Cid)
	data.Description = types.StringValue(*list.List.Description)
	data.ListItemCount = types.Int64Value(*list.List.ListItemCount)
	data.Name = types.StringValue(list.List.Name)
	data.Purpose = types.StringValue(*list.List.Purpose)
	data.Uri = types.StringValue(list.List.Uri)

	for _, item := range list.Items {
		listItemData := listItemModel{
			Did: types.StringValue(item.Subject.Did),
			Uri: types.StringValue(item.Uri),
		}

		data.Items = append(data.Items, listItemData)
	}

	for list.Cursor != nil {
		list, err := bsky.GraphGetList(ctx, d.client, *list.Cursor, 50, uri)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read List",
				err.Error(),
			)
			return
		}

		for _, item := range list.Items {
			listItemData := listItemModel{
				Did: types.StringValue(item.Subject.Did),
				Uri: types.StringValue(item.Uri),
			}

			data.Items = append(data.Items, listItemData)
		}
	}

	// Set state
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *listDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}
