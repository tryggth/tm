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

package get

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tryggth/tm/cmd/describe"
	"github.com/tryggth/tm/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func cmdListRevision(clientset *client.ConfigSet) *cobra.Command {
	return &cobra.Command{
		Use:     "revision",
		Aliases: []string{"revisions"},
		Short:   "List of knative revision resources",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				output, err := Revisions(clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(output)
			} else {
				output, err := describe.Revision(args[0], clientset)
				if err != nil {
					log.Fatalln(err)
				}
				fmt.Println(string(output))
			}
		},
	}
}

// Revisions returns list of knative revision objects
func Revisions(clientset *client.ConfigSet) (string, error) {
	list, err := clientset.Serving.ServingV1alpha1().Revisions(clientset.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if output == "" {
		table.AddRow("NAMESPACE", "REVISION")
		for _, item := range list.Items {
			table.AddRow(item.Namespace, item.Name)
		}
		return table.String(), err
	}
	return encode(list)
}
