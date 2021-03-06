/*
Copyright 2019 The Volcano Authors.

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

package queue

import (
	"fmt"
	"os"
	"strings"

	busv1alpha1 "volcano.sh/volcano/pkg/apis/bus/v1alpha1"
	"volcano.sh/volcano/pkg/apis/helpers"
	schedulingv1alpha2 "volcano.sh/volcano/pkg/apis/scheduling/v1alpha2"
	"volcano.sh/volcano/pkg/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	// Initialize client auth plugin.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func buildConfig(master, kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(master, kubeconfig)
}

func createQueueCommand(config *rest.Config, action schedulingv1alpha2.QueueAction) error {
	queueClient := versioned.NewForConfigOrDie(config)
	queue, err := queueClient.SchedulingV1alpha2().Queues().Get(operateQueueFlags.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	ctrlRef := metav1.NewControllerRef(queue, helpers.V1alpha2QueueKind)
	cmd := &busv1alpha1.Command{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-",
				queue.Name, strings.ToLower(string(action))),
			OwnerReferences: []metav1.OwnerReference{
				*ctrlRef,
			},
		},
		TargetObject: ctrlRef,
		Action:       string(action),
	}

	if _, err := queueClient.BusV1alpha1().Commands("default").Create(cmd); err != nil {
		return err
	}

	return nil
}
