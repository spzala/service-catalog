/*
Copyright 2016 The Kubernetes Authors.

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

package controller

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient"
	"github.com/kubernetes-incubator/service-catalog/pkg/controller/apiclient/mem"
	"k8s.io/kubernetes/pkg/api"
)

const (
	namespace    = "testNS"
	brokerName   = "testBroker"
	svcClassName = "testSvcClass"
	instanceName = "testInstance"
	bindingName  = "testBinding"
)

var (
	binding = servicecatalog.Binding{
		ObjectMeta: api.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
		},
		Spec: servicecatalog.BindingSpec{
			InstanceRef: api.ObjectReference{
				Namespace: namespace,
				Name:      instanceName,
			},
		},
	}
)

func makeTraversableAPIClient() apiclient.APIClient {
	instances := map[string]apiclient.InstanceClient{
		namespace: mem.NewPopulatedInstanceClient(
			map[string]*servicecatalog.Instance{
				instanceName: {
					ObjectMeta: api.ObjectMeta{
						Namespace: namespace,
						Name:      instanceName,
					},
					Spec: servicecatalog.InstanceSpec{
						ServiceClassName: svcClassName,
					},
				},
			},
		),
	}
	bindings := map[string]apiclient.BindingClient{
		namespace: mem.NewPopulatedBindingClient(
			map[string]*servicecatalog.Binding{
				bindingName: &binding,
			},
		),
	}
	return mem.NewPopulatedAPIClient(
		[]string{namespace},
		map[string]*servicecatalog.Broker{
			brokerName: {
				ObjectMeta: api.ObjectMeta{
					Namespace: namespace,
					Name:      brokerName,
				},
			},
		},
		map[string]*servicecatalog.ServiceClass{
			svcClassName: {
				ObjectMeta: api.ObjectMeta{
					Namespace: namespace,
					Name:      svcClassName,
				},
				BrokerName: brokerName,
			},
		},
		instances,
		bindings,
	)
}

func TestAll(t *testing.T) {
	storage := makeTraversableAPIClient()
	inst, err := instanceForBinding(storage, &binding)
	if err != nil {
		t.Fatalf("error getting instance for binding (%s)", err)
	}
	svcClass, err := serviceClassForInstance(storage, inst)
	if err != nil {
		t.Fatalf("error getting service class for instance (%s)", err)
	}
	broker, err := brokerForServiceClass(storage, svcClass)
	if err != nil {
		t.Fatalf("error getting broker for service class (%s)", err)
	}
	if broker == nil {
		t.Fatalf("broker was nil")
	}
}
