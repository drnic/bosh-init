// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/cloudfoundry/bosh-micro-cli/release/set (interfaces: Resolver)

package mocks

import (
	gomock "code.google.com/p/gomock/gomock"
	release "github.com/cloudfoundry/bosh-micro-cli/release"
	manifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
)

// Mock of Resolver interface
type MockResolver struct {
	ctrl     *gomock.Controller
	recorder *_MockResolverRecorder
}

// Recorder for MockResolver (not exported)
type _MockResolverRecorder struct {
	mock *MockResolver
}

func NewMockResolver(ctrl *gomock.Controller) *MockResolver {
	mock := &MockResolver{ctrl: ctrl}
	mock.recorder = &_MockResolverRecorder{mock}
	return mock
}

func (_m *MockResolver) EXPECT() *_MockResolverRecorder {
	return _m.recorder
}

func (_m *MockResolver) Filter(_param0 []manifest.ReleaseRef) error {
	ret := _m.ctrl.Call(_m, "Filter", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockResolverRecorder) Filter(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Filter", arg0)
}

func (_m *MockResolver) Find(_param0 string) (release.Release, error) {
	ret := _m.ctrl.Call(_m, "Find", _param0)
	ret0, _ := ret[0].(release.Release)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockResolverRecorder) Find(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Find", arg0)
}
