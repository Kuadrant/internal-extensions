package v1alpha1

import (
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	extctrl "github.com/kuadrant/kuadrant-operator/pkg/extension/controller"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type PipelinePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelinePolicySpec   `json:"spec,omitempty"`
	Status PipelinePolicyStatus `json:"status,omitempty"`
}

type PipelinePolicySpec struct {
	// +kubebuilder:validation:XValidation:rule="self.group == 'gateway.networking.k8s.io'",message="Invalid targetRef.group. The only supported value is 'gateway.networking.k8s.io'"
	// +kubebuilder:validation:XValidation:rule="self.kind == 'HTTPRoute' || self.kind == 'Gateway'",message="Invalid targetRef.kind. The only supported values are 'HTTPRoute' and 'Gateway'"
	TargetRef gatewayapiv1alpha2.LocalPolicyTargetReferenceWithSectionName `json:"targetRef"`

	// +optional
	ActionMethods []ActionMethodSpec `json:"actionMethods,omitempty"`

	// +optional
	Request []RequestActionSpec `json:"request,omitempty"`

	// +optional
	Response []ResponseActionSpec `json:"response,omitempty"`
}

type ActionMethodSpec struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Service         string `json:"service"`
	Method          string `json:"method"`
	MessageTemplate string `json:"messageTemplate"`
}

// +kubebuilder:validation:Enum=allow;grpc_method
type RequestActionType string

const (
	RequestActionTypeAllow      RequestActionType = "allow"
	RequestActionTypeGRPCMethod RequestActionType = "grpc_method"
)

type RequestActionSpec struct {
	Type RequestActionType `json:"type"`

	// +optional
	Predicate string `json:"predicate,omitempty"`

	// +optional
	Intention string `json:"intention,omitempty"`

	// +optional
	Method string `json:"method,omitempty"`

	// +optional
	Var string `json:"var,omitempty"`
}

// +kubebuilder:validation:Enum=add_headers;with_response_code
type ResponseActionType string

const (
	ResponseActionTypeAddHeaders       ResponseActionType = "add_headers"
	ResponseActionTypeWithResponseCode ResponseActionType = "with_response_code"
)

type ResponseActionSpec struct {
	Type ResponseActionType `json:"type"`

	// +optional
	Predicate string `json:"predicate,omitempty"`

	// +optional
	HeadersToAdd string `json:"headersToAdd,omitempty"`

	// +optional
	ResponseCode int `json:"responseCode,omitempty"`
}

func (p *PipelinePolicy) GetName() string {
	return p.Name
}

func (p *PipelinePolicy) GetNamespace() string {
	return p.Namespace
}

func (p *PipelinePolicy) GetTargetRefs() []gatewayapiv1alpha2.LocalPolicyTargetReferenceWithSectionName {
	return []gatewayapiv1alpha2.LocalPolicyTargetReferenceWithSectionName{
		p.Spec.TargetRef,
	}
}

type PipelinePolicyStatus struct {
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

func (s *PipelinePolicyStatus) Equals(other *PipelinePolicyStatus, logger logr.Logger) bool {
	if s.ObservedGeneration != other.ObservedGeneration {
		diff := cmp.Diff(s.ObservedGeneration, other.ObservedGeneration)
		logger.V(1).Info("status observedGeneration not equal", "difference", diff)
		return false
	}

	currentMarshaledJSON, _ := extctrl.ConditionMarshal(s.Conditions)
	otherMarshaledJSON, _ := extctrl.ConditionMarshal(other.Conditions)
	if string(currentMarshaledJSON) != string(otherMarshaledJSON) {
		diff := cmp.Diff(string(currentMarshaledJSON), string(otherMarshaledJSON))
		logger.V(1).Info("status conditions not equal", "difference", diff)
		return false
	}

	return true
}

// +kubebuilder:object:root=true

type PipelinePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelinePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelinePolicy{}, &PipelinePolicyList{})
}
