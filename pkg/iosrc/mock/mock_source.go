// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/brimsec/zq/pkg/iosrc (interfaces: Source)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	iosrc "github.com/brimsec/zq/pkg/iosrc"
	gomock "github.com/golang/mock/gomock"
	io "io"
	reflect "reflect"
)

// MockSource is a mock of Source interface
type MockSource struct {
	ctrl     *gomock.Controller
	recorder *MockSourceMockRecorder
}

// MockSourceMockRecorder is the mock recorder for MockSource
type MockSourceMockRecorder struct {
	mock *MockSource
}

// NewMockSource creates a new mock instance
func NewMockSource(ctrl *gomock.Controller) *MockSource {
	mock := &MockSource{ctrl: ctrl}
	mock.recorder = &MockSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSource) EXPECT() *MockSourceMockRecorder {
	return m.recorder
}

// Exists mocks base method
func (m *MockSource) Exists(arg0 context.Context, arg1 iosrc.URI) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists
func (mr *MockSourceMockRecorder) Exists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockSource)(nil).Exists), arg0, arg1)
}

// NewReader mocks base method
func (m *MockSource) NewReader(arg0 context.Context, arg1 iosrc.URI) (iosrc.Reader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewReader", arg0, arg1)
	ret0, _ := ret[0].(iosrc.Reader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewReader indicates an expected call of NewReader
func (mr *MockSourceMockRecorder) NewReader(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewReader", reflect.TypeOf((*MockSource)(nil).NewReader), arg0, arg1)
}

// NewWriter mocks base method
func (m *MockSource) NewWriter(arg0 context.Context, arg1 iosrc.URI) (io.WriteCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewWriter", arg0, arg1)
	ret0, _ := ret[0].(io.WriteCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewWriter indicates an expected call of NewWriter
func (mr *MockSourceMockRecorder) NewWriter(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewWriter", reflect.TypeOf((*MockSource)(nil).NewWriter), arg0, arg1)
}

// ReadDir mocks base method
func (m *MockSource) ReadDir(arg0 context.Context, arg1 iosrc.URI) ([]iosrc.Info, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadDir", arg0, arg1)
	ret0, _ := ret[0].([]iosrc.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadDir indicates an expected call of ReadDir
func (mr *MockSourceMockRecorder) ReadDir(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadDir", reflect.TypeOf((*MockSource)(nil).ReadDir), arg0, arg1)
}

// ReadFile mocks base method
func (m *MockSource) ReadFile(arg0 context.Context, arg1 iosrc.URI) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadFile", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadFile indicates an expected call of ReadFile
func (mr *MockSourceMockRecorder) ReadFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadFile", reflect.TypeOf((*MockSource)(nil).ReadFile), arg0, arg1)
}

// Remove mocks base method
func (m *MockSource) Remove(arg0 context.Context, arg1 iosrc.URI) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove
func (mr *MockSourceMockRecorder) Remove(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockSource)(nil).Remove), arg0, arg1)
}

// RemoveAll mocks base method
func (m *MockSource) RemoveAll(arg0 context.Context, arg1 iosrc.URI) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveAll", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveAll indicates an expected call of RemoveAll
func (mr *MockSourceMockRecorder) RemoveAll(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveAll", reflect.TypeOf((*MockSource)(nil).RemoveAll), arg0, arg1)
}

// Stat mocks base method
func (m *MockSource) Stat(arg0 context.Context, arg1 iosrc.URI) (iosrc.Info, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stat", arg0, arg1)
	ret0, _ := ret[0].(iosrc.Info)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stat indicates an expected call of Stat
func (mr *MockSourceMockRecorder) Stat(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stat", reflect.TypeOf((*MockSource)(nil).Stat), arg0, arg1)
}

// WriteFile mocks base method
func (m *MockSource) WriteFile(arg0 context.Context, arg1 []byte, arg2 iosrc.URI) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteFile", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteFile indicates an expected call of WriteFile
func (mr *MockSourceMockRecorder) WriteFile(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteFile", reflect.TypeOf((*MockSource)(nil).WriteFile), arg0, arg1, arg2)
}
