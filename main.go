package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), NewAptProvider, providerserver.ServeOpts{
		Address: "registry.terraform.io/ericmjalbert/apt",
	})
	if err != nil {
		log.Fatal(err)
	}
}
