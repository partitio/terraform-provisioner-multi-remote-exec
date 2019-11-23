# Multi Remote-Exec for Terraform

Works with Terraform 0.12.x

## General overview
The purpose of the provisioner is to provide an easy method for running dynamically 
the build-in remote-exec.

```hcl
locals {
  remote_execs = [
    {
      type: "inline"
      values: [
        "echo 'I will fail'",
        "fail"
      ],
      continue_on_failure: true,
    },
    {
      type: "inline"
      values: [
        "echo 'I did continue anyway'"
      ],
      continue_on_failure: false,
    },
    {
      type: "scripts"
      values: [
        "test.sh",
      ],
      continue_on_failure: false,
    },
  ]
}

resource "null_resource" "test" {
  connection {
    type     = "ssh"
    user     = "root"
    password = "root"
    port     = 2222
    host     = "localhost"
  }
  provisioner "multi-remote-exec" {
    dynamic "remote_exec" {
      for_each = local.remote_execs
      content {
        type   = remote_exec.value.type
        values = remote_exec.value.values
        continue_on_failure = remote_exec.value.continue_on_failure
      }
    }
  }
}

```
