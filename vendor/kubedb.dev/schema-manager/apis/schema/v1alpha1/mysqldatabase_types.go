/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MySQLDatabaseSpec defines the desired state of MySQLDatabase
type MySQLDatabaseSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MySQLDatabase. Edit mysqldatabase_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// MySQLDatabaseStatus defines the observed state of MySQLDatabase
type MySQLDatabaseStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MySQLDatabase is the Schema for the mysqldatabases API
type MySQLDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLDatabaseSpec   `json:"spec,omitempty"`
	Status MySQLDatabaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MySQLDatabaseList contains a list of MySQLDatabase
type MySQLDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQLDatabase{}, &MySQLDatabaseList{})
}
