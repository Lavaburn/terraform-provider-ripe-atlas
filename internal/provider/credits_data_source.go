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
var _ datasource.DataSource = &CreditsDataSource{}
var _ datasource.DataSourceWithConfigure = &CreditsDataSource{}

func NewCreditsDataSource() datasource.DataSource {
	return &CreditsDataSource{}
}

// ExampleDataSource defines the data source implementation.
type CreditsDataSource struct {
	client *atlas.Client
}

// ExampleDataSourceModel describes the data source data model.
type CreditsDataSourceModel struct {
	EstimatedDailyIncome      types.Int64 `tfsdk:"estimated_daily_income"`
	EstimatedDailyExpenditure types.Int64 `tfsdk:"estimated_daily_expenditure"`
	EstimatedDailyBalance     types.Int64 `tfsdk:"estimated_daily_balance"`
	CurrentBalance            types.Int64 `tfsdk:"current_balance"`
	EstimatedRunoutSeconds    types.Int64 `tfsdk:"estimated_runout_seconds"`
}

func (d *CreditsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credits"
}

func (d *CreditsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "RIPE Atlas Credits",

		Attributes: map[string]schema.Attribute{
			"current_balance": schema.Int64Attribute{
				Computed: true,
			},
			"estimated_daily_income": schema.Int64Attribute{
				Computed: true,
			},
			"estimated_daily_expenditure": schema.Int64Attribute{
				Computed: true,
			},
			"estimated_daily_balance": schema.Int64Attribute{
				Computed: true,
			},
			"estimated_runout_seconds": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *CreditsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CreditsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read Terraform data into the model
	var data CreditsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch data from API
	credits, err := d.client.GetCredits()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get credits from RIPE Atlas",
			err.Error(),
		)
		return
	}

	ctx = tflog.SetField(ctx, "credits", credits)
	tflog.Info(ctx, "RIPE Atlas credits")

	data = CreditsDataSourceModel{
		EstimatedDailyIncome:      types.Int64Value(int64(credits.EstimatedDailyIncome)),
		EstimatedDailyExpenditure: types.Int64Value(int64(credits.EstimatedDailyExpenditure)),
		EstimatedDailyBalance:     types.Int64Value(int64(credits.EstimatedDailyBalance)),
		CurrentBalance:            types.Int64Value(int64(credits.CurrentBalance)),
		EstimatedRunoutSeconds:    types.Int64Value(int64(credits.EstimatedRunoutSeconds)),
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
