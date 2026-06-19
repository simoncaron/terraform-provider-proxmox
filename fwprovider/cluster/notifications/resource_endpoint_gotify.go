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

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ resource.Resource                = &endpointGotifyResource{}
	_ resource.ResourceWithConfigure   = &endpointGotifyResource{}
	_ resource.ResourceWithImportState = &endpointGotifyResource{}
)

type endpointGotifyResource struct {
	client *notifications.Client
}

// NewEndpointGotifyResource creates a new gotify notification endpoint resource.
func NewEndpointGotifyResource() resource.Resource {
	return &endpointGotifyResource{}
}

func (r *endpointGotifyResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notifications_endpoint_gotify"
}

func (r *endpointGotifyResource) Configure(
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

func (r *endpointGotifyResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a Gotify notification endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"name": schema.StringAttribute{
				Description: "The unique name of the Gotify endpoint.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server": schema.StringAttribute{
				Description: "The Gotify server URL.",
				Required:    true,
			},
			"token": schema.StringAttribute{
				Description: "The Gotify API token. This value is write-only and will not be read back from the API.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
		},
	}
}

func (r *endpointGotifyResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state endpointGotifyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.GetGotify(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Unable to Read Gotify Endpoint", err.Error())

		return
	}

	readModel := &endpointGotifyModel{}
	readModel.fromAPI(state.ID.ValueString(), data)
	readModel.Token = state.Token

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointGotifyResource) Create( //nolint:dupl
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan endpointGotifyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateGotify(ctx, plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Gotify Endpoint", err.Error())

		return
	}

	data, err := r.client.GetGotify(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Gotify Endpoint After Creation", err.Error())

		return
	}

	readModel := &endpointGotifyModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)
	readModel.Token = plan.Token

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointGotifyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan endpointGotifyModel

	var state endpointGotifyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var toDelete []string

	attribute.CheckDelete(plan.Comment, state.Comment, &toDelete, "comment")

	reqData := plan.toAPI()
	reqData.Delete = toDelete

	err := r.client.UpdateGotify(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Gotify Endpoint", err.Error())

		return
	}

	data, err := r.client.GetGotify(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Gotify Endpoint After Update", err.Error())

		return
	}

	readModel := &endpointGotifyModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)
	readModel.Token = plan.Token

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointGotifyResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state endpointGotifyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGotify(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, api.ErrResourceDoesNotExist) {
		resp.Diagnostics.AddError("Unable to Delete Gotify Endpoint", err.Error())
	}
}

func (r *endpointGotifyResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	data, err := r.client.GetGotify(ctx, req.ID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				"Gotify Endpoint Not Found",
				fmt.Sprintf("Gotify endpoint %q was not found", req.ID),
			)

			return
		}

		resp.Diagnostics.AddError("Unable to Import Gotify Endpoint", err.Error())

		return
	}

	readModel := &endpointGotifyModel{}
	readModel.fromAPI(req.ID, data)
	// Token cannot be imported — it is write-only.

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
