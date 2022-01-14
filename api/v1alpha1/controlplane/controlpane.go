package controlplane

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ControlPlane struct {
	metav1.TypeMeta `json:",inline"`
	Spec            ControlPlaneSpec `json:"spec,omitempty"`
}

type ControlPlaneSpec struct {
	MP *ManagementPlaneSettings `protobuf:"bytes,100,opt,name=management_plane,json=managementPlane,proto3" json:"managementPlane,omitempty"`

	TM *ControlPlaneSpec_TelemetryStore_Elastic `protobuf:"bytes,200,opt,name=telemetry_store,json=telemetryStore,proto3" json:"telemetryStore,omitempty"`
}

type ControlPlaneSpec_TelemetryStore_Elastic struct {
	Elastic *ElasticSearchSettings `protobuf:"bytes,1,opt,name=elastic,json=elastic,proto3,oneof" json:"elastic,omitempty"`
}

type ElasticSearchSettings struct {
	// Elasticsearch host address (can be hostname or IP address).
	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// Port Elasticsearch is listening on.
	Port int32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// Protocol to communicate with Elasticsearch, defaults to https.
	Protocol int32 `protobuf:"varint,3,opt,name=protocol,proto3,enum=tetrateio.api.install.controlplane.v1alpha1.ElasticSearchSettings_Protocol" json:"protocol,omitempty"`
	// Use Self-Signed certificates. The Self-signed CA bundle and key must be in a secret called es-certs.
	SelfSigned bool `protobuf:"varint,4,opt,name=selfSigned,proto3" json:"selfSigned,omitempty"`
	// Major version of the Elasticsearch cluster.
	// Currently supported Elasticsearch major versions are `6` and `7`
	Version int32 `protobuf:"varint,5,opt,name=version,proto3" json:"version,omitempty"`
	// Protocol ElasticSearchSettings_Protocol `protobuf:"varint,3,opt,name=protocol,proto3,enum=tetrateio.api.install.controlplane.v1alpha1.ElasticSearc    hSettings_Protocol" json:"protocol,omitempty"`
}

type ManagementPlaneSettings struct {
	// Management plane host address (can be hostname or IP address).
	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	// Port management plane is listening on.
	Port int32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	// The unique identifier for this cluster that was created in the management plane.
	ClusterName string `protobuf:"bytes,10,opt,name=cluster_name,json=clusterName,proto3" json:"clusterName,omitempty"`
}
