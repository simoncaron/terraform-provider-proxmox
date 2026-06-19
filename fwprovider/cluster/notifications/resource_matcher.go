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

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.Resource                = &matcherResource{}
	_ resource.ResourceWithConfigure   = &matcherResource{}
	_ resource.ResourceWithImportState = &matcherResource{}
)

type matcherResource struct {
	client *notifications.Client
}

// NewMatcherResource creates a new notification matcher resource.
func NewMatcherResource() resource.Resource {
	return &matcherResource{}
}

func (r *matcherResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_notifications_matcher"
}

func (r *matcherResource) Configure(
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

func (r *matcherResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages a notification matcher.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"name": schema.StringAttribute{
				Description: "The unique name of the matcher.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Description: "Comment.",
				Optional:    true,
			},
			"disable": schema.BoolAttribute{
				Description: "Disable this matcher. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"invert_match": schema.BoolAttribute{
				Description: "Invert the match of the whole matcher.",
				Optional:    true,
			},
			"match_calendar": schema.ListAttribute{
				Description: "Match notification timestamps using systemd calendar event notation.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"match_field": schema.ListNestedAttribute{
				Description: "Match notification metadata fields.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Match type. One of `regex`, `exact`.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("regex", "exact"),
							},
						},
						"field": schema.StringAttribute{
							Description: "Metadata field name to match.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value to match against.",
							Required:    true,
						},
					},
				},
			},
			"match_severity": schema.ListAttribute{
				Description: "Match notification severities.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("unknown", "info", "notice", "warning", "error"),
					),
				},
			},
			"mode": schema.StringAttribute{
				Description: "How to combine multiple match conditions. One of `all`, `any`. Defaults to `all`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("all"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "any"),
				},
			},
			"target": schema.ListAttribute{
				Description: "List of notification target names to route matched notifications to.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *matcherResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state matcherModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data, err := r.client.GetMatcher(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Unable to Read Notification Matcher", err.Error())

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(state.ID.ValueString(), data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *matcherResource) Create( //nolint:dupl
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan matcherModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqData := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateMatcher(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Create Notification Matcher", err.Error())

		return
	}

	data, err := r.client.GetMatcher(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Notification Matcher After Creation", err.Error())

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *matcherResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan matcherModel

	var state matcherModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var toDelete []string

	attribute.CheckDelete(plan.Comment, state.Comment, &toDelete, "comment")
	attribute.CheckDelete(plan.InvertMatch, state.InvertMatch, &toDelete, "invert-match")
	checkListDelete(plan.MatchCalendar, state.MatchCalendar, &toDelete, "match-calendar")
	checkListDelete(plan.MatchSeverity, state.MatchSeverity, &toDelete, "match-severity")
	checkListDelete(plan.Target, state.Target, &toDelete, "target")

	if len(plan.MatchField) == 0 && len(state.MatchField) > 0 {
		toDelete = append(toDelete, "match-field")
	}

	reqData := plan.toAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	reqData.Delete = toDelete

	err := r.client.UpdateMatcher(ctx, reqData)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Update Notification Matcher", err.Error())

		return
	}

	data, err := r.client.GetMatcher(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Notification Matcher After Update", err.Error())

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(plan.Name.ValueString(), data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}

func (r *matcherResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state matcherModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMatcher(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, api.ErrResourceDoesNotExist) {
		resp.Diagnostics.AddError("Unable to Delete Notification Matcher", err.Error())
	}
}

func (r *matcherResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	data, err := r.client.GetMatcher(ctx, req.ID)
	if err != nil {
		if errors.Is(err, api.ErrResourceDoesNotExist) {
			resp.Diagnostics.AddError(
				"Notification Matcher Not Found",
				fmt.Sprintf("Notification matcher %q was not found", req.ID),
			)

			return
		}

		resp.Diagnostics.AddError("Unable to Import Notification Matcher", err.Error())

		return
	}

	readModel := &matcherModel{}
	readModel.fromAPI(req.ID, data)

	resp.Diagnostics.Append(resp.State.Set(ctx, readModel)...)
}
