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

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ resource.Resource                = &endpointSendmailResource{}
	_ resource.ResourceWithConfigure   = &endpointSendmailResource{}
	_ resource.ResourceWithImportState = &endpointSendmailResource{}
)

type endpointSendmailResource struct {
	client *notifications.Client
}

// NewEndpointSendmailResource creates a new sendmail notification endpoint resource.
func NewEndpointSendmailResource() resource.Resource {
	return &endpointSendmailResource{}
}

func (r *endpointSendmailResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notifications_endpoint_sendmail"
}

func (r *endpointSendmailResource) Configure(
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

func (r *endpointSendmailResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a sendmail notification endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"name": schema.StringAttribute{
				Description: "The unique name of the sendmail endpoint.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"from_address": schema.StringAttribute{
				Description: "From address for the mail.",
				Optional:    true,
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
		},
	}
}

func (r *endpointSendmailResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state endpointSendmailModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.GetSendmail(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Unable to Read Sendmail Endpoint", err.Error())

		return
	}

	readModel := &endpointSendmailModel{}
	readModel.fromAPI(state.ID.ValueString(), data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointSendmailResource) Create( //nolint:dupl
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan endpointSendmailModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqData := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateSendmail(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Sendmail Endpoint", err.Error())

		return
	}

	data, err := r.client.GetSendmail(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Sendmail Endpoint After Creation", err.Error())

		return
	}

	readModel := &endpointSendmailModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointSendmailResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan endpointSendmailModel

	var state endpointSendmailModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var toDelete []string

	attribute.CheckDelete(plan.Author, state.Author, &toDelete, "author")
	attribute.CheckDelete(plan.Comment, state.Comment, &toDelete, "comment")
	attribute.CheckDelete(plan.FromAddress, state.FromAddress, &toDelete, "from-address")
	checkListDelete(plan.Mailto, state.Mailto, &toDelete, "mailto")
	checkListDelete(plan.MailtoUser, state.MailtoUser, &toDelete, "mailto-user")

	reqData := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	reqData.Delete = toDelete

	err := r.client.UpdateSendmail(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Sendmail Endpoint", err.Error())

		return
	}

	data, err := r.client.GetSendmail(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Sendmail Endpoint After Update", err.Error())

		return
	}

	readModel := &endpointSendmailModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointSendmailResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state endpointSendmailModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSendmail(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, api.ErrResourceDoesNotExist) {
		resp.Diagnostics.AddError("Unable to Delete Sendmail Endpoint", err.Error())
	}
}

func (r *endpointSendmailResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	data, err := r.client.GetSendmail(ctx, req.ID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				"Sendmail Endpoint Not Found",
				fmt.Sprintf("Sendmail endpoint %q was not found", req.ID),
			)

			return
		}

		resp.Diagnostics.AddError("Unable to Import Sendmail Endpoint", err.Error())

		return
	}

	readModel := &endpointSendmailModel{}
	readModel.fromAPI(req.ID, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
