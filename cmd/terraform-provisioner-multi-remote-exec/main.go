package main

import (
	"github.com/hashicorp/terraform/plugin"

	"gitlab.bertha.cloud/adphi/terraform-provisioner-multi-remote-exec"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: multiremoteexec.Provisioner,
	})
}
