// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	//"net/http"

	"github.com/keltia/ripe-atlas" // PR https://github.com/keltia/ripe-atlas/pull/13

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MeasurementDataSource{}
var _ datasource.DataSourceWithConfigure = &MeasurementDataSource{}

func NewMeasurementDataSource() datasource.DataSource {
	return &MeasurementDataSource{}
}

// ExampleDataSource defines the data source implementation.
type MeasurementDataSource struct {
	client *atlas.Client
}

// ExampleDataSourceModel describes the data source data model.
type MeasurementDataSourceModel struct {
	Measurements []MeasurementsModel `tfsdk:"measurements"`
	Hidden       types.Bool          `tfsdk:"hidden"`
}

type MeasurementsModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Target      types.String `tfsdk:"target"`
	// Ping specific ?
	Interval types.Int64 `tfsdk:"interval"`
	Packets  types.Int64 `tfsdk:"packets"`
	Size     types.Int64 `tfsdk:"size"`
	// Status (not config)
	Status types.String    `tfsdk:"status"`
	Probes ProbeCountModel `tfsdk:"probes"`
}

type ProbeCountModel struct {
	Requested types.Int64 `tfsdk:"requested"`
	Scheduled types.Int64 `tfsdk:"scheduled"`
}

func (d *MeasurementDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_measurement"
}

func (d *MeasurementDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "RIPE Atlas Measurement",

		Attributes: map[string]schema.Attribute{
			"hidden": schema.BoolAttribute{
				Optional: true,
			},
			"measurements": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"target": schema.StringAttribute{
							Computed: true,
						},
						"interval": schema.Int64Attribute{
							Computed: true,
						},
						"packets": schema.Int64Attribute{
							Computed: true,
						},
						"size": schema.Int64Attribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"probes": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"requested": schema.Int64Attribute{
									Computed: true,
								},
								"scheduled": schema.Int64Attribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *MeasurementDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*atlas.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *atlas.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *MeasurementDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read Terraform data into the model
	var data MeasurementDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch data from API
	request := make(map[string]string)
	request["mine"] = "true"
	if data.Hidden.ValueBool() {
		request["hidden"] = "true"
	}

	measurements, err := d.client.GetMeasurements(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get measurements from RIPE Atlas",
			err.Error(),
		)
		return
	}

	for _, measurement := range measurements {
		ctx = tflog.SetField(ctx, "measurement", measurement)
		tflog.Info(ctx, "RIPE Atlas measurement")

		m := MeasurementsModel{
			ID:          types.Int64Value(int64(measurement.ID)),
			Description: types.StringValue(measurement.Description),
			Type:        types.StringValue(measurement.Type),
			Target:      types.StringValue(measurement.Target),
			// Ping specific ?
			Interval: types.Int64Value(int64(measurement.Interval)),
			Packets:  types.Int64Value(int64(measurement.Packets)),
			Size:     types.Int64Value(int64(measurement.Size)),
			// Status (not config)
			Status: types.StringValue(measurement.Status.Name),
			Probes: ProbeCountModel{
				Requested: types.Int64Value(int64(measurement.ProbesRequested)),
				Scheduled: types.Int64Value(int64(measurement.ProbesScheduled)),
			},
			// Other: af/creation_time/is_oneoff/packet_interval/start_time/stop_time
			// Other: port/protocol/
		}
		data.Measurements = append(data.Measurements, m)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
