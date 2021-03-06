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

package server

import (
	"flag"
	"io"
	"os"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/service-catalog/pkg/registry/servicecatalog/server"
	"github.com/spf13/cobra"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_5"
	genericserveroptions "k8s.io/kubernetes/pkg/genericapiserver/options"
)

const (
	// Store generated SSL certificates in a place that won't collide with the
	// k8s core API server.
	certDirectory = "/var/run/kubernetes-service-catalog"

	// I made this up to match some existing paths. I am not sure if there
	// are any restrictions on the format or structure beyond text
	// separated by slashes.
	etcdPathPrefix = "/k8s.io/service-catalog"

	// GroupName I made this up. Maybe we'll need it.
	GroupName = "service-catalog.k8s.io"

	storageTypeFlagName    = "storageType"
	tprGlobalNamespaceName = "tprGlobalNamespace"
)

// NewCommandServer creates a new cobra command to run our server.
func NewCommandServer(
	out io.Writer,
	clIface clientset.Interface,
) *cobra.Command {
	// Create the command that runs the API server
	cmd := &cobra.Command{
		Short: "run a service-catalog server",
	}
	// We pass flags object to sub option structs to have them configure
	// themselves. Each options adds its own command line flags
	// in addition to the flags that are defined above.
	flags := cmd.Flags()
	flags.AddGoFlagSet(flag.CommandLine)

	opts := &ServiceCatalogServerOptions{
		GenericServerRunOptions: genericserveroptions.NewServerRunOptions(),
		SecureServingOptions:    genericserveroptions.NewSecureServingOptions(),
		AuthenticationOptions:   genericserveroptions.NewDelegatingAuthenticationOptions(),
		AuthorizationOptions:    genericserveroptions.NewDelegatingAuthorizationOptions(),
		InsecureServingOptions:  genericserveroptions.NewInsecureServingOptions(),
		EtcdOptions:             NewEtcdOptions(),
		TPROptions:              NewTPROptions(),
	}
	opts.addFlags(flags)
	// Set generated SSL cert path correctly
	opts.SecureServingOptions.ServerCert.CertDirectory = certDirectory

	flags.Parse(os.Args[1:])

	storageType, err := opts.StorageType()
	if err != nil {
		glog.Fatalf("invalid storage type '%s' (%s)", storageType, err)
		return nil
	}
	if storageType == server.StorageTypeEtcd {
		glog.Infof("using etcd for storage")
		opts.EtcdOptions.cl = clIface
		// Store resources in etcd under our special prefix
		opts.EtcdOptions.StorageConfig.Prefix = etcdPathPrefix
	} else {
		glog.Infof("using third party resources for storage")
		opts.TPROptions.defaultGlobalNamespace = "servicecatalog"
		opts.TPROptions.clIface = clIface
	}

	cmd.Run = func(c *cobra.Command, args []string) {
		if err := RunServer(opts); err != nil {
			glog.Fatalf("error running server (%s)", err)
			return
		}
	}

	return cmd
}
