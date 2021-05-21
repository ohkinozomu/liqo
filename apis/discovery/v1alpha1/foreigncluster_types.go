/*


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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	advtypes "github.com/liqotech/liqo/apis/sharing/v1alpha1"
	crdclient "github.com/liqotech/liqo/pkg/crdClient"
	"github.com/liqotech/liqo/pkg/discovery"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PeeringPhaseType indicates the phase of a peering with a remote cluster.
type PeeringPhaseType string

const (
	// PeeringPhaseNone indicates that there is no peering.
	PeeringPhaseNone PeeringPhaseType = "None"
	// PeeringPhasePending indicates that the peering is pending,
	// and we are waiting fore the remote cluster feedback.
	PeeringPhasePending PeeringPhaseType = "Pending"
	// PeeringPhaseEstablished indicates that the peering has been established.
	PeeringPhaseEstablished PeeringPhaseType = "Established"
	// PeeringPhaseDisconnecting indicates that the peering is being deleted.
	PeeringPhaseDisconnecting PeeringPhaseType = "Disconnecting"
)

// ForeignClusterSpec defines the desired state of ForeignCluster.
type ForeignClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foreign Cluster Identity.
	ClusterIdentity ClusterIdentity `json:"clusterIdentity,omitempty"`
	// Namespace where Liqo is deployed. (Deprecated)
	Namespace string `json:"namespace,omitempty"`
	// Enable join process to foreign cluster.
	// +kubebuilder:default=false
	Join bool `json:"join,omitempty"`
	// +kubebuilder:validation:Enum="LAN";"WAN";"Manual";"IncomingPeering"
	// +kubebuilder:default="Manual"
	// How this ForeignCluster has been discovered.
	DiscoveryType discovery.Type `json:"discoveryType,omitempty"`
	// URL where to contact foreign Auth service.
	AuthURL string `json:"authUrl"`
	// +kubebuilder:validation:Enum="Unknown";"Trusted";"Untrusted"
	// +kubebuilder:default="Unknown"
	// Indicates if this remote cluster is trusted or not.
	TrustMode discovery.TrustMode `json:"trustMode,omitempty"`
	// If discoveryType is LAN or WAN and this indicates the number of seconds after that
	// this ForeignCluster will be removed if no updates have been received.
	// +kubebuilder:validation:Minimum=0
	TTL int `json:"ttl,omitempty"`
}

// ClusterIdentity contains the information about a remote cluster (ID and Name).
type ClusterIdentity struct {
	// Foreign Cluster ID, this is a unique identifier of that cluster.
	ClusterID string `json:"clusterID"`
	// Foreign Cluster Name to be shown in GUIs.
	ClusterName string `json:"clusterName,omitempty"`
}

// ForeignClusterStatus defines the observed state of ForeignCluster.
type ForeignClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TenantControlNamespaces names in the peered clusters
	// +kubebuilder:validation:Optional
	TenantControlNamespace TenantControlNamespace `json:"tenantControlNamespace"`

	Outgoing Outgoing `json:"outgoing,omitempty"`
	Incoming Incoming `json:"incoming,omitempty"`
	// It stores most important network statuses.
	Network Network `json:"network,omitempty"`
	// Authentication status.
	// +kubebuilder:validation:Enum="Pending";"Accepted";"Refused";"EmptyRefused"
	// +kubebuilder:default="Pending"
	AuthStatus discovery.AuthStatus `json:"authStatus,omitempty"`
}

// TenantControlNamespace contains the names of the local and the remote
// namespaces assigned to the pair of clusters.
type TenantControlNamespace struct {
	// local TenantNamespace name
	Local string `json:"local,omitempty"`
	// remote TenantNamespace name
	Remote string `json:"remote,omitempty"`
}

// ResourceLink contains information on the reference of an kubernetes resource.
type ResourceLink struct {
	// Indicates if the resource is available.
	Available bool `json:"available"`
	// Object Reference to the resource.
	Reference *v1.ObjectReference `json:"reference,omitempty"`
}

// Network contains the information on the network status.
type Network struct {
	// Local NetworkConfig link.
	LocalNetworkConfig ResourceLink `json:"localNetworkConfig,omitempty"`
	// Remote NetworkConfig link.
	RemoteNetworkConfig ResourceLink `json:"remoteNetworkConfig,omitempty"`
	// TunnelEndpoint link.
	TunnelEndpoint ResourceLink `json:"tunnelEndpoint,omitempty"`
}

// Outgoing contains the status of the outgoing peering.
type Outgoing struct {
	// Indicates if peering request has been created and this remote cluster is sharing its resources to us.
	// +kubebuilder:validation:Enum="None";"Pending";"Established";"Disconnecting"
	// +kubebuilder:default="None"
	PeeringPhase PeeringPhaseType `json:"joinPhase,omitempty"`
	// Name of created PR. (Deprecated)
	RemotePeeringRequestName string `json:"remote-peering-request-name,omitempty"`
	// Object Reference to created Advertisement CR. (Deprecated)
	Advertisement *v1.ObjectReference `json:"advertisement,omitempty"`
	// Indicates if related identity is available. (Deprecated)
	AvailableIdentity bool `json:"availableIdentity,omitempty"`
	// Object reference to related identity. (Deprecated)
	IdentityRef *v1.ObjectReference `json:"identityRef,omitempty"`
	// Advertisement status. (Deprecated)
	AdvertisementStatus advtypes.AdvPhase `json:"advertisementStatus,omitempty"`
}

// Incoming contains the status of the incoming peering.
type Incoming struct {
	// Indicates if peering request has been created and this remote cluster is using our local resources.
	// +kubebuilder:validation:Enum="None";"Pending";"Established";"Disconnecting"
	// +kubebuilder:default="None"
	PeeringPhase PeeringPhaseType `json:"joinPhase,omitempty"`
	// Object Reference to created PeeringRequest CR. (Deprecated)
	PeeringRequest *v1.ObjectReference `json:"peeringRequest,omitempty"`
	// Indicates if related identity is available. (Deprecated)
	AvailableIdentity bool `json:"availableIdentity,omitempty"`
	// Object reference to related identity. (Deprecated)
	IdentityRef *v1.ObjectReference `json:"identityRef,omitempty"`
	// Status of Advertisement created from this PeeringRequest. (Deprecated)
	AdvertisementStatus advtypes.AdvPhase `json:"advertisementStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// ForeignCluster is the Schema for the foreignclusters API.
// +kubebuilder:printcolumn:name="Outgoing peering phase",type=string,JSONPath=`.status.outgoing.joinPhase`
// +kubebuilder:printcolumn:name="Incoming peering phase",type=string,JSONPath=`.status.incoming.joinPhase`
// +kubebuilder:printcolumn:name="Authentication status",type=string,JSONPath=`.status.authStatus`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ForeignCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ForeignClusterSpec   `json:"spec,omitempty"`
	Status ForeignClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ForeignClusterList contains a list of ForeignCluster.
type ForeignClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ForeignCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ForeignCluster{}, &ForeignClusterList{})

	if err := AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	crdclient.AddToRegistry("foreignclusters", &ForeignCluster{}, &ForeignClusterList{}, nil, ForeignClusterGroupResource)
}
