package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &accountResource{}
	_ resource.ResourceWithConfigure   = &accountResource{}
	_ resource.ResourceWithImportState = &accountResource{}
	_ resource.ResourceWithModifyPlan  = &accountResource{}
)

// NewAccountResource is a helper function to simplify the provider implementation.
func NewAccountResource() resource.Resource {
	return &accountResource{}
}

// accountResource is the resource implementation.
type accountResource struct {
	client          *xrpc.Client
	anonymousClient *xrpc.Client
}

type accountResourceModel struct {
	Did      types.String `tfsdk:"did"`
	Email    types.String `tfsdk:"email"`
	Handle   types.String `tfsdk:"handle"`
	Password types.String `tfsdk:"password"`
	// TODO to support account import:
	//recoveryKey     types.String `tfsdk:"recovery_key"`

	// These don't make sense to manage via TF:
	//inviteCode
	//verificationCode
	//verificationPhone
}

// Metadata returns the resource type name.
func (l *accountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

// Schema defines the schema for the resource.
func (r *accountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Accounts. This resource requires the provider to be configured with the `pds_admin_password `.",
		Attributes: map[string]schema.Attribute{
			"did": schema.StringAttribute{
				MarkdownDescription: "Account's DID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the account",
				Optional:            true,
			},
			"handle": schema.StringAttribute{
				MarkdownDescription: "Requested handle for the account",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Set the initial account password on create or update the password for an existing account. If not specified on create, a password will be generated and included in the Terraform output in plaintext.",
				Sensitive:           true,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (l *accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from a plan.
	var plan accountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	password := plan.Password.ValueString()
	if password == "" {
		generatedPassword, err := getRandomPassword()
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating account",
				"Failed to generate random initial password: "+err.Error(),
			)
			return
		}
		password = generatedPassword
	}

	// Create an invite code
	createInviteCodeInput := &atproto.ServerCreateInviteCode_Input{
		UseCount: 1,
	}
	inviteCode, err := atproto.ServerCreateInviteCode(ctx, l.client, createInviteCodeInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating account",
			"Could not create invite code, unexpected error: "+err.Error(),
		)
		return
	}

	// Generate API request body from plan. Adapted from the account migration script:
	// https://github.com/bluesky-social/indigo/blob/main/cmd/goat/account_migrate.go
	createRecordInput := atproto.ServerCreateAccount_Input{
		Handle:     plan.Handle.ValueString(),
		Email:      plan.Email.ValueStringPointer(),
		Password:   &password,
		InviteCode: &inviteCode.Code,
	}

	// Create new account.
	createOutput, err := atproto.ServerCreateAccount(ctx, l.anonymousClient, &createRecordInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating account",
			"Could not create account, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values.
	plan.Did = types.StringValue(createOutput.Did)

	// Set state to fully populated data.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Password.ValueString() == "" {
		resp.Diagnostics.AddWarning(
			"Initial password created",
			"Generated initial password for account "+plan.Handle.ValueString()+": "+password,
		)
	}
}

// Read refreshes the Terraform state with the latest data.
func (l *accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state accountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := atproto.AdminGetAccountInfo(ctx, l.client, state.Did.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to retrieve account",
			"Could not retrieve the account, error: "+err.Error(),
		)
		return
	}

	state.Handle = types.StringValue(account.Handle)
	state.Email = types.StringValue(*account.Email)

	// Set refreshed state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (l *accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from a plan.
	var plan accountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	// Get current state.
	var state accountResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update email
	if !strings.EqualFold(plan.Email.ValueString(), state.Email.ValueString()) {
		updateEmailInput := &atproto.AdminUpdateAccountEmail_Input{
			Account: state.Did.ValueString(),
			Email:   plan.Email.ValueString(),
		}
		err := atproto.AdminUpdateAccountEmail(ctx, l.client, updateEmailInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating account",
				"Could not update account email, error: "+err.Error(),
			)
			return
		}
		state.Email = plan.Email
	}

	// update handle
	if !strings.EqualFold(plan.Handle.ValueString(), state.Handle.ValueString()) {
		updateHandleInput := &atproto.AdminUpdateAccountHandle_Input{
			Did:    state.Did.ValueString(),
			Handle: plan.Handle.ValueString(),
		}
		err := atproto.AdminUpdateAccountHandle(ctx, l.client, updateHandleInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating account",
				"Could not update account handle, error: "+err.Error(),
			)
		}
		state.Handle = plan.Handle
	}

	// update password
	if !strings.EqualFold(plan.Password.ValueString(), state.Password.ValueString()) {
		updatePasswordInput := &atproto.AdminUpdateAccountPassword_Input{
			Did:      state.Did.ValueString(),
			Password: plan.Password.ValueString(),
		}
		err := atproto.AdminUpdateAccountPassword(ctx, l.client, updatePasswordInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating account",
				"Could not update account password, error: "+err.Error(),
			)
			return
		}
		state.Password = plan.Password
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (l *accountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state accountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteRequest := &atproto.AdminDeleteAccount_Input{
		Did: state.Did.ValueString(),
	}
	err := atproto.AdminDeleteAccount(ctx, l.client, deleteRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting account",
			"Could not delete account, error: "+err.Error(),
		)
	}
}

// Configure adds the provider configured client to the resource.
func (l *accountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	if client.AdminToken == nil {
		resp.Diagnostics.AddError(
			"PDSAdminPassword required",
			"An admin token is required to manage accounts, please configure the provider with the PDSAdminPassword.",
		)
		return
	}

	// Make a copy of the client without any Auth set to force the client to use the admin token from the Headers for all account requests.
	// https://github.com/bluesky-social/indigo/issues/994
	l.client = &xrpc.Client{
		Host:      client.Host,
		UserAgent: client.UserAgent,
		Headers: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:"+*client.AdminToken)),
		},
		AdminToken: client.AdminToken,
		Client:     client.Client,
		Auth:       nil,
	}

	// Make yet another copy of the client, this one without even an Auth header set,
	// because the PDS doesn't expect account creations from an invite to be authenticated.
	l.anonymousClient = &xrpc.Client{
		Host:       client.Host,
		UserAgent:  client.UserAgent,
		Headers:    map[string]string{},
		AdminToken: client.AdminToken,
		Client:     client.Client,
		Auth:       nil,
	}
}

func (l *accountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import DID and save to did attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("did"), req, resp)
}

func (l *accountResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan accountResourceModel
	if !req.Plan.Raw.IsNull() {
		diags := req.Plan.Get(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		// warn if a plaintext password will be generated during account creation
		if req.State.Raw.IsNull() && plan.Password.ValueString() == "" {
			resp.Diagnostics.AddWarning(
				"Initial password not specified",
				"Initial password for account "+plan.Handle.ValueString()+" was not specified, one will be generated and included in the Terraform output in plaintext.",
			)
		}
	}
}

func getRandomPassword() (string, error) {
	// generate a password similar to how pdsadmin does it: https://github.com/bluesky-social/pds/blob/f054eefea58e6cddf17eda14a55ecf157c2e034e/pdsadmin/account.sh#L65
	length := 30
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	password := base64.URLEncoding.EncodeToString(bytes)
	if len(password) > length {
		password = password[:length]
	}
	return password, nil
}
