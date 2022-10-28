// Code generated by mockery v2.10.0. DO NOT EDIT.

package mocks

import (
	mgo "github.com/TykTechnologies/tyk-pump/pumps/internal/mgo"
	mock "github.com/stretchr/testify/mock"
)

// DatabaseManager is an autogenerated mock type for the DatabaseManager type
type DatabaseManager struct {
	mock.Mock
}

// C provides a mock function with given fields: name
func (_m *DatabaseManager) C(name string) mgo.CollectionManager {
	ret := _m.Called(name)

	var r0 mgo.CollectionManager
	if rf, ok := ret.Get(0).(func(string) mgo.CollectionManager); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(mgo.CollectionManager)
		}
	}

	return r0
}

// CollectionNames provides a mock function with given fields:
func (_m *DatabaseManager) CollectionNames() ([]string, error) {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DropDatabase provides a mock function with given fields:
func (_m *DatabaseManager) DropDatabase() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Run provides a mock function with given fields: cmd, result
func (_m *DatabaseManager) Run(cmd interface{}, result interface{}) error {
	ret := _m.Called(cmd, result)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}, interface{}) error); ok {
		r0 = rf(cmd, result)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
