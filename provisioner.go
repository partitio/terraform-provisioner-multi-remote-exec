package multiremoteexec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-linereader"
)

func Provisioner() terraform.ResourceProvisioner {
	return &schema.Provisioner{
		Schema: map[string]*schema.Schema{
			"remote_exec": remoteExecSchema,
		},
		ApplyFunc:    provision,
		ValidateFunc: nil,
	}
}

func provision(ctx context.Context) error {
	connState := ctx.Value(schema.ProvRawStateKey).(*terraform.InstanceState)
	data := ctx.Value(schema.ProvConfigDataKey).(*schema.ResourceData)
	o := ctx.Value(schema.ProvOutputKey).(terraform.UIOutput)

	// Get a new communicator
	comm, err := communicator.New(connState)
	if err != nil {
		return err
	}

	_, ok := data.GetOk("remote_exec")
	if !ok {
		o.Output("No remote_exec found")
		return nil
	}
	execs, err := parse(data)
	if err != nil {
		return err
	}
	for _, e := range execs {
		if err := e.Collect(); err != nil {
			return err
		}
		if err := runScripts(ctx, o, comm, e.Scripts); err != nil && !e.ContinueOnFailure {
			return err
		}
	}
	return nil
}

func parse(data *schema.ResourceData) ([]RemoteExec, error) {
	d, ok := data.Get("remote_exec").([]interface{})
	if !ok {
		return nil, errors.New("no remote_exec found")
	}
	var execs []RemoteExec
	for _, v := range d {
		js, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse remote_exec element: %v", err)
		}
		e := RemoteExec{}
		if err := json.Unmarshal(js, &e); err != nil {
			return nil, fmt.Errorf("failed to parse remote_exec element: %v", err)
		}
		execs = append(execs, e)
	}
	return execs, nil
}

// runScripts is used to copy and execute a set of scripts
func runScripts(
	ctx context.Context,
	o terraform.UIOutput,
	comm communicator.Communicator,
	scripts []io.ReadCloser) error {

	retryCtx, cancel := context.WithTimeout(ctx, comm.Timeout())
	defer cancel()

	// Wait and retry until we establish the connection
	err := communicator.Retry(retryCtx, func() error {
		return comm.Connect(o)
	})
	if err != nil {
		return err
	}

	// Wait for the context to end and then disconnect
	go func() {
		<-ctx.Done()
		comm.Disconnect()
	}()

	for _, script := range scripts {
		var cmd *remote.Cmd

		outR, outW := io.Pipe()
		errR, errW := io.Pipe()
		defer outW.Close()
		defer errW.Close()

		go copyOutput(o, outR)
		go copyOutput(o, errR)

		remotePath := comm.ScriptPath()

		if err := comm.UploadScript(remotePath, script); err != nil {
			return fmt.Errorf("Failed to upload script: %v", err)
		}

		cmd = &remote.Cmd{
			Command: remotePath,
			Stdout:  outW,
			Stderr:  errW,
		}
		if err := comm.Start(cmd); err != nil {
			return fmt.Errorf("Error starting script: %v", err)
		}

		if err := cmd.Wait(); err != nil {
			return err
		}

		// Upload a blank follow up file in the same path to prevent residual
		// script contents from remaining on remote machine
		empty := bytes.NewReader([]byte(""))
		if err := comm.Upload(remotePath, empty); err != nil {
			// This feature is best-effort.
			log.Printf("[WARN] Failed to upload empty follow up script: %v", err)
		}
	}

	return nil
}

func copyOutput(
	o terraform.UIOutput, r io.Reader) {
	lr := linereader.New(r)
	for line := range lr.Ch {
		o.Output(line)
	}
}
