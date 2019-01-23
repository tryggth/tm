/*
Copyright (c) 2018 TriggerMesh, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/mholt/archiver"
	"github.com/tryggth/tm/pkg/client"
	"k8s.io/client-go/tools/remotecommand"
)

// Copy contains information to copy local path to remote destination
type Copy struct {
	Pod         string
	Container   string
	Source      string
	Destination string
}

var (
	uploadPath = "/tmp/tm/upload"
	command    = "tar -xvf -"
)

// Upload receives Copy structure, creates tarball of local source path and uploads it to active (un)tar process on remote pod
func (c *Copy) Upload(clientset *client.ConfigSet) error {
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return err
	}

	tar := path.Join(uploadPath, randString(10))
	if err := archiver.Tar.Make(tar, []string{c.Source}); err != nil {
		return err
	}

	fileReader, err := os.Open(tar)
	if err != nil {
		return err
	}

	if c.Destination != "" {
		command = fmt.Sprintf("%s -C %s", command, c.Destination)
	}

	stdout, stderr, err := c.RemoteExec(clientset, command, fileReader)
	if err != nil {
		fmt.Printf("Stdout: %s\nStderr: %s\nErr: %s\n", stdout, stderr, err)
		return err
	}
	return nil
}

// RemoteExec executes command on remote pod and returns stdout and stderr output
func (c *Copy) RemoteExec(clientset *client.ConfigSet, command string, file io.Reader) (string, string, error) {
	var commandLine string
	for _, v := range strings.Fields(command) {
		commandLine = fmt.Sprintf("%s&command=%s", commandLine, v)
	}
	if c.Container != "" {
		commandLine = fmt.Sprintf("&container=%s%s", c.Container, commandLine)
	}
	stdin := "false"
	if file != nil {
		stdin = "true"
	}
	// workaround to form correct URL
	urlAndParams := strings.Split(clientset.Core.RESTClient().Post().URL().String(), "?")
	url := fmt.Sprintf("%sapi/v1/namespaces/%s/pods/%s/exec?stderr=true&stdin=%s&stdout=true%s", urlAndParams[0], clientset.Namespace, c.Pod, stdin, commandLine)
	if len(urlAndParams) == 2 {
		url = fmt.Sprintf("%s&%s", url, urlAndParams[1])
	}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", "", err
	}

	exec, err := remotecommand.NewSPDYExecutor(clientset.Config, "POST", req.URL)
	if err != nil {
		return "", "", err
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  file,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}
