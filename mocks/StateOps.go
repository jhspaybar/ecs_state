package mocks

import "github.com/jhspaybar/ecsstate"
import "github.com/stretchr/testify/mock"

import "github.com/jinzhu/gorm"

type StateOps struct {
	mock.Mock
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

// DB provides a mock function with given fields:
func (_m *StateOps) DB() *gorm.DB {
	ret := _m.Called()

	var r0 *gorm.DB
	if rf, ok := ret.Get(0).(func() *gorm.DB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorm.DB)
		}
	}

	return r0
}
