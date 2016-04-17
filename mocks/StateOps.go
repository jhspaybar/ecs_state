package mocks

import "github.com/jhspaybar/ecsstate"
import "github.com/stretchr/testify/mock"

import "github.com/aws/aws-sdk-go/service/ecs"

type StateOps struct {
	mock.Mock
}

// Initialize provides a mock function with given fields: clusterName, ecs, logger
func (_m *StateOps) Initialize(clusterName string, e *ecs.ECS, logger ecsstate.Logger) *ecsstate.State {
	ret := _m.Called(clusterName, e, logger)

	var r0 *ecsstate.State
	if rf, ok := ret.Get(0).(func(string, *ecs.ECS, ecsstate.Logger) *ecsstate.State); ok {
		r0 = rf(clusterName, e, logger)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecsstate.State)
		}
	}

	return r0
}

// FindLocationsForTaskDefinition provides a mock function with given fields: td
func (_m *StateOps) FindLocationsForTaskDefinition(td string) *[]ecsstate.ContainerInstance {
	ret := _m.Called(td)

	var r0 *[]ecsstate.ContainerInstance
	if rf, ok := ret.Get(0).(func(string) *[]ecsstate.ContainerInstance); ok {
		r0 = rf(td)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]ecsstate.ContainerInstance)
		}
	}

	return r0
}

// FindTaskDefinition provides a mock function with given fields: td
func (_m *StateOps) FindTaskDefinition(td string) ecsstate.TaskDefinition {
	ret := _m.Called(td)

	var r0 ecsstate.TaskDefinition
	if rf, ok := ret.Get(0).(func(string) ecsstate.TaskDefinition); ok {
		r0 = rf(td)
	} else {
		r0 = ret.Get(0).(ecsstate.TaskDefinition)
	}

	return r0
}

// RefreshClusterState provides a mock function with given fields:
func (_m *StateOps) RefreshClusterState() {
	_m.Called()
}

// RefreshContainerInstanceState provides a mock function with given fields:
func (_m *StateOps) RefreshContainerInstanceState() {
	_m.Called()
}

// RefreshTaskState provides a mock function with given fields:
func (_m *StateOps) RefreshTaskState() {
	_m.Called()
}
