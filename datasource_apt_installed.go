package main

import (
	"bufio"
	"context"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &aptInstalledDataSource{}

type aptInstalledDataSource struct{}

type aptInstalledModel struct {
	Packages []aptInstalledPackage `tfsdk:"packages"`
}

type aptInstalledPackage struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

func NewAptInstalledDataSource() datasource.DataSource {
	return &aptInstalledDataSource{}
}

func (d *aptInstalledDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_installed"
}

func (d *aptInstalledDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all installed apt packages for audit purposes.",
		Attributes: map[string]schema.Attribute{
			"packages": schema.ListNestedAttribute{
				Description: "All installed apt packages.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Package name.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Installed version.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *aptInstalledDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Package}\t${Version}\n")
	out, err := cmd.Output()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list packages", err.Error())
		return
	}

	var packages []aptInstalledPackage
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		packages = append(packages, aptInstalledPackage{
			Name:    types.StringValue(parts[0]),
			Version: types.StringValue(parts[1]),
		})
	}

	state := aptInstalledModel{Packages: packages}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
