package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &aptProvider{}

type aptProvider struct{}

func NewAptProvider() provider.Provider {
	return &aptProvider{}
}

func (p *aptProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apt"
}

func (p *aptProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage apt packages on Ubuntu/Debian systems.",
	}
}

func (p *aptProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *aptProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAptPackageResource,
	}
}

func (p *aptProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAptInstalledDataSource,
	}
}
