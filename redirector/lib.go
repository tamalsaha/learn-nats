package redirector

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type REST struct {
	client.Client
	proxyTransport http.RoundTripper
}

var (
	_ rest.Redirector = &REST{}
)

// ResourceLocation returns a URL to which one can send traffic for the specified service.
func (r *REST) ResourceLocation(ctx context.Context, id string) (*url.URL, http.RoundTripper, error) {
	// Allow ID as "svcname", "svcname:port", or "scheme:svcname:port".
	svcScheme, svcName, portStr, valid := utilnet.SplitSchemeNamePort(id)
	if !valid {
		return nil, nil, errors.NewBadRequest(fmt.Sprintf("invalid service request %q", id))
	}
	objKey := client.ObjectKey{Namespace: genericapirequest.NamespaceValue(ctx), Name: svcName}

	// If a port *number* was specified, find the corresponding service port name
	if portNum, err := strconv.ParseInt(portStr, 10, 64); err == nil {
		var svc core.Service
		err := r.Get(ctx, objKey, &svc)
		if err != nil {
			return nil, nil, err
		}
		// svc := obj.(*core.Service)
		found := false
		for _, svcPort := range svc.Spec.Ports {
			if int64(svcPort.Port) == portNum {
				// use the declared port's name
				portStr = svcPort.Name
				found = true
				break
			}
		}
		if !found {
			return nil, nil, errors.NewServiceUnavailable(fmt.Sprintf("no service port %d found for service %q", portNum, svcName))
		}
	}

	var eps core.Endpoints
	err := r.Get(ctx, objKey, &eps)
	if err != nil {
		return nil, nil, err
	}
	if len(eps.Subsets) == 0 {
		return nil, nil, errors.NewServiceUnavailable(fmt.Sprintf("no endpoints available for service %q", svcName))
	}
	// Pick a random Subset to start searching from.
	ssSeed := rand.Intn(len(eps.Subsets))
	// Find a Subset that has the port.
	for ssi := 0; ssi < len(eps.Subsets); ssi++ {
		ss := &eps.Subsets[(ssSeed+ssi)%len(eps.Subsets)]
		if len(ss.Addresses) == 0 {
			continue
		}
		for i := range ss.Ports {
			if ss.Ports[i].Name == portStr {
				addrSeed := rand.Intn(len(ss.Addresses))
				// This is a little wonky, but it's expensive to test for the presence of a Pod
				// So we repeatedly try at random and validate it, this means that for an invalid
				// service with a lot of endpoints we're going to potentially make a lot of calls,
				// but in the expected case we'll only make one.
				for try := 0; try < len(ss.Addresses); try++ {
					addr := ss.Addresses[(addrSeed+try)%len(ss.Addresses)]
					// TODO(thockin): do we really need this check?
					if err := r.isValidAddress(ctx, &addr); err != nil {
						utilruntime.HandleError(fmt.Errorf("address %v isn't valid (%v)", addr, err))
						continue
					}
					ip := addr.IP
					port := int(ss.Ports[i].Port)
					return &url.URL{
						Scheme: svcScheme,
						Host:   net.JoinHostPort(ip, strconv.Itoa(port)),
					}, r.proxyTransport, nil
				}
				utilruntime.HandleError(fmt.Errorf("failed to find a valid address, skipping subset: %v", ss))
			}
		}
	}
	return nil, nil, errors.NewServiceUnavailable(fmt.Sprintf("no endpoints available for service %q", id))
}

func (r *REST) isValidAddress(ctx context.Context, addr *core.EndpointAddress) error {
	if addr.TargetRef == nil {
		return fmt.Errorf("address has no target ref, skipping: %v", addr)
	}
	if genericapirequest.NamespaceValue(ctx) != addr.TargetRef.Namespace {
		return fmt.Errorf("address namespace doesn't match context namespace")
	}
	var pod core.Pod
	err := r.Get(ctx, client.ObjectKey{Namespace: "", Name: addr.TargetRef.Name}, &pod)
	if err != nil {
		return err
	}
	for _, podIP := range pod.Status.PodIPs {
		if podIP.IP == addr.IP {
			return nil
		}
	}
	return fmt.Errorf("pod ip(s) doesn't match endpoint ip, skipping: %v vs %s (%s/%s)", pod.Status.PodIPs, addr.IP, addr.TargetRef.Namespace, addr.TargetRef.Name)
}
