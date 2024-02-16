// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

// Package aws is a generated GoMock package.
package aws

import (
	reflect "reflect"

	rds "github.com/aws/aws-sdk-go/service/rds"
	gomock "github.com/golang/mock/gomock"
)

// MockRDSClient is a mock of RDSClient interface.
type MockRDSClient struct {
	ctrl     *gomock.Controller
	recorder *MockRDSClientMockRecorder
}

// MockRDSClientMockRecorder is the mock recorder for MockRDSClient.
type MockRDSClientMockRecorder struct {
	mock *MockRDSClient
}

// NewMockRDSClient creates a new mock instance.
func NewMockRDSClient(ctrl *gomock.Controller) *MockRDSClient {
	mock := &MockRDSClient{ctrl: ctrl}
	mock.recorder = &MockRDSClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRDSClient) EXPECT() *MockRDSClientMockRecorder {
	return m.recorder
}

// GetAuroraClusterEndpoints mocks base method.
func (m *MockRDSClient) GetAuroraClusterEndpoints(dbClusterIdentifiers []string) (map[string]*AuroraCluster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuroraClusterEndpoints", dbClusterIdentifiers)
	ret0, _ := ret[0].(map[string]*AuroraCluster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAuroraClusterEndpoints indicates an expected call of GetAuroraClusterEndpoints.
func (mr *MockRDSClientMockRecorder) GetAuroraClusterEndpoints(dbClusterIdentifiers interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuroraClusterEndpoints", reflect.TypeOf((*MockRDSClient)(nil).GetAuroraClusterEndpoints), dbClusterIdentifiers)
}

// MockrdsService is a mock of rdsService interface.
type MockrdsService struct {
	ctrl     *gomock.Controller
	recorder *MockrdsServiceMockRecorder
}

// MockrdsServiceMockRecorder is the mock recorder for MockrdsService.
type MockrdsServiceMockRecorder struct {
	mock *MockrdsService
}

// NewMockrdsService creates a new mock instance.
func NewMockrdsService(ctrl *gomock.Controller) *MockrdsService {
	mock := &MockrdsService{ctrl: ctrl}
	mock.recorder = &MockrdsServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockrdsService) EXPECT() *MockrdsServiceMockRecorder {
	return m.recorder
}

// DescribeDBInstances mocks base method.
func (m *MockrdsService) DescribeDBInstances(input *rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeDBInstances", input)
	ret0, _ := ret[0].(*rds.DescribeDBInstancesOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeDBInstances indicates an expected call of DescribeDBInstances.
func (mr *MockrdsServiceMockRecorder) DescribeDBInstances(input interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeDBInstances", reflect.TypeOf((*MockrdsService)(nil).DescribeDBInstances), input)
}
