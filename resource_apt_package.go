package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &aptPackageResource{}

type aptPackageResource struct{}

type aptPackageModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
	Ensure  types.String `tfsdk:"ensure"`
}

func NewAptPackageResource() resource.Resource {
	return &aptPackageResource{}
}

func (r *aptPackageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (r *aptPackageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an apt package.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The apt package name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Description: "The installed version (read from dpkg-query).",
				Computed:    true,
			},
			"ensure": schema.StringAttribute{
				Description: "Whether the package should be \"present\" or \"absent\".",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("present"),
			},
		},
	}
}

func (r *aptPackageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aptPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	if plan.Ensure.ValueString() == "present" {
		if err := aptInstall(name); err != nil {
			resp.Diagnostics.AddError("Install failed", err.Error())
			return
		}
	}

	version, err := dpkgQueryVersion(name)
	if err != nil {
		resp.Diagnostics.AddError("Version query failed", err.Error())
		return
	}

	plan.Version = types.StringValue(version)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aptPackageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state aptPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	version, err := dpkgQueryVersion(name)
	if err != nil {
		// Package not installed — remove from state so it gets recreated
		resp.State.RemoveResource(ctx)
		return
	}

	state.Version = types.StringValue(version)
	state.Ensure = types.StringValue("present")
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aptPackageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan aptPackageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	switch plan.Ensure.ValueString() {
	case "absent":
		if err := aptRemove(name); err != nil {
			resp.Diagnostics.AddError("Remove failed", err.Error())
			return
		}
		plan.Version = types.StringValue("")
	case "present":
		if err := aptInstall(name); err != nil {
			resp.Diagnostics.AddError("Install failed", err.Error())
			return
		}
		version, err := dpkgQueryVersion(name)
		if err != nil {
			resp.Diagnostics.AddError("Version query failed", err.Error())
			return
		}
		plan.Version = types.StringValue(version)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aptPackageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aptPackageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := aptRemove(state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Remove failed", err.Error())
	}
}

func aptInstall(name string) error {
	cmd := exec.Command("sudo", "flock", "/var/lib/dpkg/lock-frontend",
		"apt-get", "install", "-y", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt-get install %s: %w\n%s", name, err, string(out))
	}
	return nil
}

func aptRemove(name string) error {
	cmd := exec.Command("sudo", "flock", "/var/lib/dpkg/lock-frontend",
		"apt-get", "remove", "-y", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt-get remove %s: %w\n%s", name, err, string(out))
	}
	return nil
}

func dpkgQueryVersion(name string) (string, error) {
	cmd := exec.Command("dpkg-query", "-W", "-f=${Version}", name)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("dpkg-query %s: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}
