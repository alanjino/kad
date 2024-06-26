package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TektonProjectStatus string

const (
	TektonProjectAvailable            TektonProjectStatus = "available"
	TektonProjectConfigured           TektonProjectStatus = "configured"
	TektonProjectConfigurationOngoing TektonProjectStatus = "configuration-ongoing"
	TektonProjectConfigurationFailed  TektonProjectStatus = "configuration-failed"
)

const (
	TektonPipelineConfigUseCase = "tekton"
	TektonHostName              = "tekton"
	TektonProjectSync           = "tekton-project-sync"
)

type TektonProject struct {
	Id             string `json:"id,omitempty"`
	GitProjectId   string `json:"git_project_id,omitempty"`
	GitProjectUrl  string `json:"git_project_url,omitempty"`
	Status         string `json:"status,omitempty"`
	LastUpdateTime string `json:"last_update_time,omitempty"`
	WorkflowId     string `json:"workflow_id,omitempty"`
	WorkflowStatus string `json:"workflow_status,omitempty"`
}

type TektonPipeline struct {
	Id                     string   `json:"id,omitempty"`
	PipelineName           string   `json:"pipeline_name,omitempty"`
	WebhookURL             string   `json:"webhook_url,omitempty"`
	GitOrgId               string   `json:"git_org_id,omitempty"`
	GitOrgUrl              string   `json:"git_org_url,omitempty"`
	ContainerRegId         []string `json:"container_reg_id,omitempty"`
	ManagedClusterId       string   `json:"managed_cluster_id,omitempty"`
	CrossplaneGitProjectId string   `json:"crossplane_git_project_id,omitempty"`
	Status                 string   `json:"status,omitempty"`
	LastUpdateTime         string   `json:"last_update_time,omitempty"`
	WorkflowId             string   `json:"workflow_id,omitempty"`
	WorkflowStatus         string   `json:"workflow_status,omitempty"`
}

type EventListenerStatus struct {
	ConditionedStatus `json:",inline"`
}

type EventListener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status EventListenerStatus `json:"status,omitempty"`
}

type EventListeners struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventListener `json:"items"`
}
