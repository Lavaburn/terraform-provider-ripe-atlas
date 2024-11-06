// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	//"net/http"

	"github.com/keltia/ripe-atlas" // PR https://github.com/keltia/ripe-atlas/pull/13

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	//"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &RipeAtlasProvider{}
var _ provider.ProviderWithFunctions = &RipeAtlasProvider{}

// ripeAtlasProvider defines the provider implementation.
type RipeAtlasProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type RipeAtlasProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
}

func (p *RipeAtlasProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ripe-atlas"
	resp.Version = p.version
}

func (p *RipeAtlasProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "RIPE Atlas API Key",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *RipeAtlasProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data RipeAtlasProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	if data.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown RIPE Atlas API Key",
			"The provider cannot create the RIPE Atlas API client as there is an unknown configuration value for the RIPE Atlas API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the RIPE_ATLAS_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	api_key := os.Getenv("RIPE_ATLAS_API_KEY")
	if !data.ApiKey.IsNull() {
		api_key = data.ApiKey.ValueString()
	}
	if api_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing RIPE Atlas API Key",
			"The provider cannot create the  RIPE Atlas API client as there is a missing or empty value for the  RIPE Atlas API Key. "+
				"Set the host value in the configuration or use the RIPE_ATLAS_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new HashiCups client using the configuration values
	client, err := atlas.NewClient(atlas.Config{
		APIKey: api_key,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create RIPE Atlas API Client",
			"An unexpected error occurred when creating the RIPE Atlas API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"RIPE Atlas Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *RipeAtlasProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMeasurementResource,
	}
}

func (p *RipeAtlasProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMeasurementDataSource,
		NewCreditsDataSource,
	}
}

func (p *RipeAtlasProvider) Functions(ctx context.Context) []func() function.Function {
	return nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RipeAtlasProvider{
			version: version,
		}
	}
}
