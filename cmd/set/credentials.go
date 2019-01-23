//    Copyright 2018 TriggerMesh, Inc
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package set

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tryggth/tm/pkg/client"
	"golang.org/x/crypto/ssh/terminal"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Credentials contains docker registry credentials
type Credentials struct {
	Host     string
	Username string
	Password string
	Pull     bool
	Push     bool
}

// SetRegistryCreds creates Secret with docker registry credentials json which later can be mounted as config.json file
func (c *Credentials) SetRegistryCreds(name string, stdin bool, clientset *client.ConfigSet) error {
	secrets := make(map[string]string)
	if stdin {
		if err := c.readStdin(); err != nil {
			return err
		}
	}
	secret := fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\"}}}", c.Host, c.Username, c.Password)
	s, err := clientset.Core.CoreV1().Secrets(clientset.Namespace).Get(name, metav1.GetOptions{})
	if err == nil {
		for k, v := range s.Data {
			secrets[k] = string(v)
		}
	}

	if c.Pull || c.Pull == c.Push {
		secrets[".dockerconfigjson"] = secret
	}
	if c.Push || c.Push == c.Pull {
		secrets["config.json"] = secret
	}
	newSecret := corev1.Secret{
		Type: "kubernetes.io/dockerconfigjson",
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: clientset.Namespace,
		},
		StringData: secrets,
	}
	if s.GetName() != "" {
		newSecret.ObjectMeta.ResourceVersion = s.ObjectMeta.ResourceVersion
		if _, err = clientset.Core.CoreV1().Secrets(clientset.Namespace).Update(&newSecret); err != nil {
			return err
		}
	} else {
		if _, err = clientset.Core.CoreV1().Secrets(clientset.Namespace).Create(&newSecret); err != nil {
			return err
		}
	}

	if c.Pull || c.Pull == c.Push {
		sa, err := clientset.Core.CoreV1().ServiceAccounts(clientset.Namespace).Get("default", metav1.GetOptions{})
		if err != nil {
			return err
		}
		sa.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: name},
		}
		if _, err := clientset.Core.CoreV1().ServiceAccounts(clientset.Namespace).Update(sa); err != nil {
			return err
		}
	}
	return nil
}

func (c *Credentials) readStdin() error {
	reader := bufio.NewReader(os.Stdin)
	if len(c.Host) == 0 {
		fmt.Printf("Registry: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		c.Host = strings.Replace(text, "\n", "", -1)
	}
	if len(c.Username) == 0 {
		fmt.Print("Username: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		c.Username = strings.Replace(text, "\n", "", -1)
	}
	if len(c.Password) == 0 {
		fmt.Print("Password: ")
		text, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println()
		c.Password = string(text)
	}
	return nil
}
