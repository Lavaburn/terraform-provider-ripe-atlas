// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"
	//"net/http"

	"github.com/keltia/ripe-atlas" // PR https://github.com/keltia/ripe-atlas/pull/13

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MeasurementResource{}
var _ resource.ResourceWithImportState = &MeasurementResource{}
var _ resource.ResourceWithConfigure = &MeasurementResource{}

func NewMeasurementResource() resource.Resource {
	return &MeasurementResource{}
}

// ExampleResource defines the resource implementation.
type MeasurementResource struct {
	client *atlas.Client
}

// ExampleResourceModel describes the resource data model.
type MeasurementResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"` // TODO: VALIDATE ENUM ?
	Target      types.String `tfsdk:"target"`
	// Ping specific ?
	Interval types.Int64 `tfsdk:"interval"`
	Packets  types.Int64 `tfsdk:"packets"`
	Size     types.Int64 `tfsdk:"size"`
	// Probes (on Create)
	ProbeSet []ProbeSetResourceModel `tfsdk:"probe_set"`
	// Terraform Internal
	LastUpdated types.String `tfsdk:"last_updated"`
}

type ProbeSetResourceModel struct {
	Number types.Int64  `tfsdk:"number"`
	Type   types.String `tfsdk:"type"` // area, country, prefix, asn, probes, msm
	Value  types.String `tfsdk:"value"`
	//TagsInclude			types.String	`tfsdk:"include"`
	//TagsExclude			types.String	`tfsdk:"exclude"`
}

func (r *MeasurementResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_measurement"
}

func (r *MeasurementResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "RIPE Atlas Measurement",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"ping", "dns", "http", "ntp", "sslcert", "traceroute"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"interval": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(300),
				Validators: []validator.Int64{
					int64validator.Between(30, 3600), // 30 sec - 1 hour ?
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"packets": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(3),
				Validators: []validator.Int64{
					int64validator.Between(1, 10), // max 10 packets ?
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(48),
				Validators: []validator.Int64{
					int64validator.Between(48, 1500), // 48 - 1500 bytes ?
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"probe_set": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"number": schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 50), // Max 50 probes per set
							},
							/*PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},*/
						},
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf([]string{"area", "country", "asn", "probes", "msm"}...),
							},
							/*PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},*/
						},
						"value": schema.StringAttribute{
							Required: true,
							/*PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},*/
						},
					},
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
				// TODO: problem ?
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *MeasurementResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*atlas.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *atlas.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *MeasurementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data MeasurementResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare creation request
	request := r.client.NewMeasurement()
	request.IsOneoff = false // TODO: PARAM ??
	//request.Times = infinite
	//request.StartTime = Now
	//request.StopTime = Never

	request.Probes = nil
	for _, ps := range data.ProbeSet {
		request.Probes = append(request.Probes, atlas.NewProbeSet(int(ps.Number.ValueInt64()), ps.Type.ValueString(), ps.Value.ValueString(), ""))
	}

	// Definitions: []Definition
	opts := map[string]string{
		"Type":        data.Type.ValueString(),
		"Description": data.Description.ValueString(),
		"AF":          "4",
		"Target":      data.Target.ValueString(),
		// Ping Only ?
		"Packets":  strconv.FormatInt(data.Packets.ValueInt64(), 10),
		"Interval": strconv.FormatInt(data.Interval.ValueInt64(), 10),
		"Size":     strconv.FormatInt(data.Size.ValueInt64(), 10),
		// PacketInterval
	}
	request.AddDefinition(opts)

	// Call API
	ctx = tflog.SetField(ctx, "request", request)
	tflog.Info(ctx, "Creating RIPE Atlas measurement")
	if data.Type.ValueString() == "ping" {
		measurements, err := r.client.Ping(request)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ping measurement, got error: %s", err))
			return
		}

		for _, newId := range measurements.Measurements {
			data.ID = types.Int64Value(int64(newId))
		}
		//TODO: DNS / HTTP / NTP / SSLCert / Traceroute
	} else {
		resp.Diagnostics.AddError("Type Error", fmt.Sprintf("Measurement type %s not supported!", data.Type.ValueString()))
		return
	}

	if data.ID.IsUnknown() {
		resp.Diagnostics.AddError("No ID Retrieved", "Error occurred while creating object. No ID retrieved!")
		return
	}

	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MeasurementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MeasurementResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsUnknown() { // TODO: IF NOT IMPORTED, ID IS UNKNOWN AND SHOULD CREATE NEW MEASUREMENT !
		return
	}

	tflog.Info(ctx, "Fetching RIPE Atlas measurement")
	measurement, err := r.client.GetMeasurement(int(data.ID.ValueInt64()), true)
	tflog.Info(ctx, "RIPE Atlas measurement fetched")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get measurement from RIPE Atlas",
			err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "measurement", measurement)
	tflog.Info(ctx, "RIPE Atlas measurement found")

	probe_set := []ProbeSetResourceModel{}
	for _, participation := range measurement.ParticipationRequests {
		probe_set = append(probe_set, ProbeSetResourceModel{
			Type:   types.StringValue(participation.Type),
			Value:  types.StringValue(participation.Value),
			Number: types.Int64Value(int64(participation.Requested)),
		})
	}

	data = MeasurementResourceModel{
		ID:          types.Int64Value(int64(measurement.ID)),
		Description: types.StringValue(measurement.Description),
		Type:        types.StringValue(measurement.Type),
		Target:      types.StringValue(measurement.Target),
		// Ping specific ?
		Interval: types.Int64Value(int64(measurement.Interval)),
		Packets:  types.Int64Value(int64(measurement.Packets)),
		Size:     types.Int64Value(int64(measurement.Size)),
		ProbeSet: probe_set,
		//            LastUpdated
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MeasurementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data MeasurementResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Description (IF...?)
	ctx = tflog.SetField(ctx, "new_description", data.Description)
	tflog.Info(ctx, "Updating description...")

	// TODO: API - PATCH {description: "MyFirstTest1"}

	// Update Probes (IF...?)
	ctx = tflog.SetField(ctx, "new_description", data.ProbeSet)
	tflog.Info(ctx, "Updating probes...")

	// TODO: API - POST ID/participation-requests {type: "country", value: "SS", requested: "1", action: "add"}

	resp.Diagnostics.AddError(
		"Update not supported!",
		"Update is currently not supported!",
	)
	return

	// tflog.Info(ctx, "Fetching RIPE Atlas measurement")
	// _, err := r.client.GetMeasurement(int(data.ID.ValueInt64()))
	// tflog.Info(ctx, "RIPE Atlas measurement fetched")
	// if err != nil {
	//   resp.Diagnostics.AddError(
	//     "Unable to get measurement from RIPE Atlas",
	//     err.Error(),
	//   )
	//   return
	// }

	//data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Save updated data into Terraform state
	//resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MeasurementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data MeasurementResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting RIPE Atlas measurement")
	err := r.client.DeleteMeasurement(int(data.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get measurement from RIPE Atlas",
			err.Error(),
		)
		return
	} else {
		tflog.Info(ctx, "RIPE Atlas measurement deleted")
	}

	// TODO: HIDE ON UI ???
}

func (r *MeasurementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing item",
			"Could not import item, unexpected error (ID should be an integer): "+err.Error(),
		)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), id)
}
