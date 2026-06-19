/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ resource.Resource                = &endpointSMTPResource{}
	_ resource.ResourceWithConfigure   = &endpointSMTPResource{}
	_ resource.ResourceWithImportState = &endpointSMTPResource{}
)

type endpointSMTPResource struct {
	client *notifications.Client
}

// NewEndpointSMTPResource creates a new SMTP notification endpoint resource.
func NewEndpointSMTPResource() resource.Resource {
	return &endpointSMTPResource{}
}

func (r *endpointSMTPResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notifications_endpoint_smtp"
}

func (r *endpointSMTPResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *proxmox.Client, got: %T", req.ProviderData),
		)

		return
	}

	r.client = cfg.Client.Cluster().Notifications()
}

func (r *endpointSMTPResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages an SMTP notification endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"name": schema.StringAttribute{
				Description: "The unique name of the SMTP endpoint.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server": schema.StringAttribute{
				Description: "The address of the SMTP server.",
				Required:    true,
			},
			"from_address": schema.StringAttribute{
				Description: "The from address for outgoing mail.",
				Required:    true,
			},
			"author": schema.StringAttribute{
				Description: "Author of the mail. Defaults to `Proxmox VE`.",
				Optional:    true,
			},
			"comment": schema.StringAttribute{
				Description: "Comment.",
				Optional:    true,
			},
			"disable": schema.BoolAttribute{
				Description: "Disable this endpoint. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"mailto": schema.ListAttribute{
				Description: "List of email recipients.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mailto_user": schema.ListAttribute{
				Description: "List of PVE users to send notifications to.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"mode": schema.StringAttribute{
				Description: "Encryption mode. One of `insecure`, `starttls`, `tls`. Defaults to `tls`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("tls"),
				Validators: []validator.String{
					stringvalidator.OneOf("insecure", "starttls", "tls"),
				},
			},
			"port": schema.Int64Attribute{
				Description: "The port to be used. Defaults to the standard port for the selected `mode`.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username for SMTP authentication.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for SMTP authentication. This value is write-only and will not be read back from the API.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *endpointSMTPResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state endpointSMTPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.GetSMTP(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Unable to Read SMTP Endpoint", err.Error())

		return
	}

	readModel := &endpointSMTPModel{}
	readModel.fromAPI(state.ID.ValueString(), data)
	readModel.Password = state.Password

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointSMTPResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan endpointSMTPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqData := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateSMTP(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create SMTP Endpoint", err.Error())

		return
	}

	data, err := r.client.GetSMTP(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SMTP Endpoint After Creation", err.Error())

		return
	}

	readModel := &endpointSMTPModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)
	readModel.Password = plan.Password

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointSMTPResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan endpointSMTPModel

	var state endpointSMTPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var toDelete []string

	attribute.CheckDelete(plan.Author, state.Author, &toDelete, "author")
	attribute.CheckDelete(plan.Comment, state.Comment, &toDelete, "comment")
	attribute.CheckDelete(plan.Port, state.Port, &toDelete, "port")
	attribute.CheckDelete(plan.Username, state.Username, &toDelete, "username")
	attribute.CheckDelete(plan.Password, state.Password, &toDelete, "password")
	checkListDelete(plan.Mailto, state.Mailto, &toDelete, "mailto")
	checkListDelete(plan.MailtoUser, state.MailtoUser, &toDelete, "mailto-user")

	reqData := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	reqData.Delete = toDelete

	err := r.client.UpdateSMTP(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update SMTP Endpoint", err.Error())

		return
	}

	data, err := r.client.GetSMTP(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read SMTP Endpoint After Update", err.Error())

		return
	}

	readModel := &endpointSMTPModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)
	readModel.Password = plan.Password

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointSMTPResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state endpointSMTPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSMTP(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, api.ErrResourceDoesNotExist) {
		resp.Diagnostics.AddError("Unable to Delete SMTP Endpoint", err.Error())
	}
}

func (r *endpointSMTPResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	data, err := r.client.GetSMTP(ctx, req.ID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				"SMTP Endpoint Not Found",
				fmt.Sprintf("SMTP endpoint %q was not found", req.ID),
			)

			return
		}

		resp.Diagnostics.AddError("Unable to Import SMTP Endpoint", err.Error())

		return
	}

	readModel := &endpointSMTPModel{}
	readModel.fromAPI(req.ID, data)
	// Password cannot be imported — it is write-only.

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
