package redirector

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	api "k8s.io/kubernetes/pkg/apis/core"
)

type REST struct {
	*genericregistry.Store
	primaryIPFamily   api.IPFamily
	secondaryIPFamily api.IPFamily
	alloc             Allocators
	endpoints         EndpointsStorage
	pods              PodStorage
	proxyTransport    http.RoundTripper
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

	// If a port *number* was specified, find the corresponding service port name
	if portNum, err := strconv.ParseInt(portStr, 10, 64); err == nil {
		obj, err := r.Get(ctx, svcName, &metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		svc := obj.(*api.Service)
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

	obj, err := r.endpoints.Get(ctx, svcName, &metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	eps := obj.(*api.Endpoints)
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
					if err := isValidAddress(ctx, &addr, r.pods); err != nil {
						utilruntime.HandleError(fmt.Errorf("Address %v isn't valid (%v)", addr, err))
						continue
					}
					ip := addr.IP
					port := int(ss.Ports[i].Port)
					return &url.URL{
						Scheme: svcScheme,
						Host:   net.JoinHostPort(ip, strconv.Itoa(port)),
					}, r.proxyTransport, nil
				}
				utilruntime.HandleError(fmt.Errorf("Failed to find a valid address, skipping subset: %v", ss))
			}
		}
	}
	return nil, nil, errors.NewServiceUnavailable(fmt.Sprintf("no endpoints available for service %q", id))
}

func isValidAddress(ctx context.Context, addr *api.EndpointAddress, pods rest.Getter) error {
	if addr.TargetRef == nil {
		return fmt.Errorf("Address has no target ref, skipping: %v", addr)
	}
	if genericapirequest.NamespaceValue(ctx) != addr.TargetRef.Namespace {
		return fmt.Errorf("Address namespace doesn't match context namespace")
	}
	obj, err := pods.Get(ctx, addr.TargetRef.Name, &metav1.GetOptions{})
	if err != nil {
		return err
	}
	pod, ok := obj.(*api.Pod)
	if !ok {
		return fmt.Errorf("failed to cast to pod: %v", obj)
	}
	if pod == nil {
		return fmt.Errorf("pod is missing, skipping (%s/%s)", addr.TargetRef.Namespace, addr.TargetRef.Name)
	}
	for _, podIP := range pod.Status.PodIPs {
		if podIP.IP == addr.IP {
			return nil
		}
	}
	return fmt.Errorf("pod ip(s) doesn't match endpoint ip, skipping: %v vs %s (%s/%s)", pod.Status.PodIPs, addr.IP, addr.TargetRef.Namespace, addr.TargetRef.Name)
}

// normalizeClusterIPs adjust clusterIPs based on ClusterIP.  This must not
// consider any other fields.
func normalizeClusterIPs(after After, before Before) {
	oldSvc, newSvc := before.Service, after.Service

	// In all cases here, we don't need to over-think the inputs.  Validation
	// will be called on the new object soon enough.  All this needs to do is
	// try to divine what user meant with these linked fields. The below
	// is verbosely written for clarity.

	// **** IMPORTANT *****
	// as a governing rule. User must (either)
	// -- Use singular only (old client)
	// -- singular and plural fields (new clients)

	if oldSvc == nil {
		// This was a create operation.
		// User specified singular and not plural (e.g. an old client), so init
		// plural for them.
		if len(newSvc.Spec.ClusterIP) > 0 && len(newSvc.Spec.ClusterIPs) == 0 {
			newSvc.Spec.ClusterIPs = []string{newSvc.Spec.ClusterIP}
			return
		}

		// we don't init singular based on plural because
		// new client must use both fields

		// Either both were not specified (will be allocated) or both were
		// specified (will be validated).
		return
	}

	// This was an update operation

	// ClusterIPs were cleared by an old client which was trying to patch
	// some field and didn't provide ClusterIPs
	if len(oldSvc.Spec.ClusterIPs) > 0 && len(newSvc.Spec.ClusterIPs) == 0 {
		// if ClusterIP is the same, then it is an old client trying to
		// patch service and didn't provide ClusterIPs
		if oldSvc.Spec.ClusterIP == newSvc.Spec.ClusterIP {
			newSvc.Spec.ClusterIPs = oldSvc.Spec.ClusterIPs
		}
	}

	// clusterIP is not the same
	if oldSvc.Spec.ClusterIP != newSvc.Spec.ClusterIP {
		// this is a client trying to clear it
		if len(oldSvc.Spec.ClusterIP) > 0 && len(newSvc.Spec.ClusterIP) == 0 {
			// if clusterIPs are the same, then clear on their behalf
			if sameClusterIPs(oldSvc, newSvc) {
				newSvc.Spec.ClusterIPs = nil
			}

			// if they provided nil, then we are fine (handled by patching case above)
			// if they changed it then validation will catch it
		} else {
			// ClusterIP has changed but not cleared *and* ClusterIPs are the same
			// then we set ClusterIPs based on ClusterIP
			if sameClusterIPs(oldSvc, newSvc) {
				newSvc.Spec.ClusterIPs = []string{newSvc.Spec.ClusterIP}
			}
		}
	}
}

// patchAllocatedValues allows clients to avoid a read-modify-write cycle while
// preserving values that we allocated on their behalf.  For example, they
// might create a Service without specifying the ClusterIP, in which case we
// allocate one.  If they resubmit that same YAML, we want it to succeed.
func patchAllocatedValues(after After, before Before) {
	oldSvc, newSvc := before.Service, after.Service

	if needsClusterIP(oldSvc) && needsClusterIP(newSvc) {
		if newSvc.Spec.ClusterIP == "" {
			newSvc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
		}
		if len(newSvc.Spec.ClusterIPs) == 0 && len(oldSvc.Spec.ClusterIPs) > 0 {
			newSvc.Spec.ClusterIPs = oldSvc.Spec.ClusterIPs
		}
	}

	if needsNodePort(oldSvc) && needsNodePort(newSvc) {
		nodePortsUsed := func(svc *api.Service) sets.Int32 {
			used := sets.NewInt32()
			for _, p := range svc.Spec.Ports {
				if p.NodePort != 0 {
					used.Insert(p.NodePort)
				}
			}
			return used
		}

		// Build a set of all the ports in oldSvc that are also in newSvc.  We know
		// we can't patch these values.
		used := nodePortsUsed(oldSvc).Intersection(nodePortsUsed(newSvc))

		// Map NodePorts by name.  The user may have changed other properties
		// of the port, but we won't see that here.
		np := map[string]int32{}
		for i := range oldSvc.Spec.Ports {
			p := &oldSvc.Spec.Ports[i]
			np[p.Name] = p.NodePort
		}

		// If newSvc is missing values, try to patch them in when we know them and
		// they haven't been used for another port.

		for i := range newSvc.Spec.Ports {
			p := &newSvc.Spec.Ports[i]
			if p.NodePort == 0 {
				oldVal := np[p.Name]
				if !used.Has(oldVal) {
					p.NodePort = oldVal
				}
			}
		}
	}

	if needsHCNodePort(oldSvc) && needsHCNodePort(newSvc) {
		if newSvc.Spec.HealthCheckNodePort == 0 {
			newSvc.Spec.HealthCheckNodePort = oldSvc.Spec.HealthCheckNodePort
		}
	}
}

func needsClusterIP(svc *api.Service) bool {
	if svc.Spec.Type == api.ServiceTypeExternalName {
		return false
	}
	return true
}

func needsNodePort(svc *api.Service) bool {
	if svc.Spec.Type == api.ServiceTypeNodePort || svc.Spec.Type == api.ServiceTypeLoadBalancer {
		return true
	}
	return false
}

func needsHCNodePort(svc *api.Service) bool {
	if svc.Spec.Type != api.ServiceTypeLoadBalancer {
		return false
	}
	if svc.Spec.ExternalTrafficPolicy != api.ServiceExternalTrafficPolicyTypeLocal {
		return false
	}
	return true
}
