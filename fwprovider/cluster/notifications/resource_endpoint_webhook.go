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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/cluster/notifications"
)

var (
	_ resource.Resource                = &endpointWebhookResource{}
	_ resource.ResourceWithConfigure   = &endpointWebhookResource{}
	_ resource.ResourceWithImportState = &endpointWebhookResource{}
)

type endpointWebhookResource struct {
	client *notifications.Client
}

// NewEndpointWebhookResource creates a new webhook notification endpoint resource.
func NewEndpointWebhookResource() resource.Resource {
	return &endpointWebhookResource{}
}

func (r *endpointWebhookResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notifications_endpoint_webhook"
}

func (r *endpointWebhookResource) Configure(
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

func (r *endpointWebhookResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a webhook notification endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"name": schema.StringAttribute{
				Description: "The unique name of the webhook endpoint.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Description: "The webhook server URL.",
				Required:    true,
			},
			"method": schema.StringAttribute{
				Description: "HTTP method to use. One of `post`, `put`, `get`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("post", "put", "get"),
				},
			},
			"body": schema.StringAttribute{
				Description: "HTTP body template. Stored as plain text, base64-encoded when sent to the API.",
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
			"header": schema.ListNestedAttribute{
				Description: "List of HTTP headers to set.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Header name.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Header value.",
							Required:    true,
						},
					},
				},
			},
			"secret": schema.ListNestedAttribute{
				Description: "List of secrets available as template variables. Values are write-only and will not be read back from the API.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Secret name.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Secret value. This value is write-only.",
							Required:    true,
							Sensitive:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *endpointWebhookResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state endpointWebhookModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.GetWebhook(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Unable to Read Webhook Endpoint", err.Error())

		return
	}

	readModel := &endpointWebhookModel{}
	readModel.fromAPI(state.ID.ValueString(), data)
	readModel.Secret = state.Secret

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointWebhookResource) Create( //nolint:dupl
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan endpointWebhookModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateWebhook(ctx, plan.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Webhook Endpoint", err.Error())

		return
	}

	data, err := r.client.GetWebhook(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Webhook Endpoint After Creation", err.Error())

		return
	}

	readModel := &endpointWebhookModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)
	readModel.Secret = plan.Secret

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointWebhookResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan endpointWebhookModel

	var state endpointWebhookModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var toDelete []string

	attribute.CheckDelete(plan.Body, state.Body, &toDelete, "body")
	attribute.CheckDelete(plan.Comment, state.Comment, &toDelete, "comment")

	if len(plan.Header) == 0 && len(state.Header) > 0 {
		toDelete = append(toDelete, "header")
	}

	if len(plan.Secret) == 0 && len(state.Secret) > 0 {
		toDelete = append(toDelete, "secret")
	}

	reqData := plan.toAPI()
	reqData.Delete = toDelete

	err := r.client.UpdateWebhook(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Webhook Endpoint", err.Error())

		return
	}

	data, err := r.client.GetWebhook(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Webhook Endpoint After Update", err.Error())

		return
	}

	readModel := &endpointWebhookModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)
	readModel.Secret = plan.Secret

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *endpointWebhookResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state endpointWebhookModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWebhook(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, api.ErrResourceDoesNotExist) {
		resp.Diagnostics.AddError("Unable to Delete Webhook Endpoint", err.Error())
	}
}

func (r *endpointWebhookResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	data, err := r.client.GetWebhook(ctx, req.ID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				"Webhook Endpoint Not Found",
				fmt.Sprintf("Webhook endpoint %q was not found", req.ID),
			)

			return
		}

		resp.Diagnostics.AddError("Unable to Import Webhook Endpoint", err.Error())

		return
	}

	readModel := &endpointWebhookModel{}
	readModel.fromAPI(req.ID, data)
	// Secret values cannot be imported — they are write-only.

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
