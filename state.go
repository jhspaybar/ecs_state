package ecsstate

import (
	"github.com/aws/aws-sdk-go/service/ecs"
)

// StateOps is the interface for refreshing and interacting with the local
// ECS state.
type StateOps interface {
	Initialize(clusterName string, ecs *ecs.ECS, logger Logger) *State
	FindLocationsForTaskDefinition(td string) *[]ContainerInstance
	FindTaskDefinition(td string) TaskDefinition
	RefreshClusterState()
	RefreshContainerInstanceState()
	RefreshTaskState()
}
