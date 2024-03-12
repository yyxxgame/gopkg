// Code generated by MockGen. DO NOT EDIT.
// Source: ./stores/elastic/elastic.go

// Package elastic is a generated GoMock package.
package elastic

import (
	"context"
	gomock "github.com/golang/mock/gomock"
	v7elastic "github.com/olivere/elastic/v7"
	"reflect"
)

// mockIEsClient is a mock of IEsClient interface.
type mockIEsClient struct {
	ctrl     *gomock.Controller
	recorder *mockIEsClientMockRecorder
}

// mockIEsClientMockRecorder is the mock recorder for mockIEsClient.
type mockIEsClientMockRecorder struct {
	mock *mockIEsClient
}

// newMockIEsClient creates a new mock instance.
func newMockIEsClient(ctrl *gomock.Controller) *mockIEsClient {
	mock := &mockIEsClient{ctrl: ctrl}
	mock.recorder = &mockIEsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *mockIEsClient) EXPECT() *mockIEsClientMockRecorder {
	return m.recorder
}

// Insert mocks base method.
func (m *mockIEsClient) Insert(chain InsertChain) (*v7elastic.IndexResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", chain)
	ret0, _ := ret[0].(*v7elastic.IndexResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Insert indicates an expected call of Insert.
func (mr *mockIEsClientMockRecorder) Insert(chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*mockIEsClient)(nil).Insert), chain)
}

// InsertCtx mocks base method.
func (m *mockIEsClient) InsertCtx(ctx context.Context, chain InsertChain) (*v7elastic.IndexResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertCtx", ctx, chain)
	ret0, _ := ret[0].(*v7elastic.IndexResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertCtx indicates an expected call of InsertCtx.
func (mr *mockIEsClientMockRecorder) InsertCtx(ctx, chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertCtx", reflect.TypeOf((*mockIEsClient)(nil).InsertCtx), ctx, chain)
}

// Query mocks base method.
func (m *mockIEsClient) Query(chain QueryChain) (*v7elastic.SearchResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", chain)
	ret0, _ := ret[0].(*v7elastic.SearchResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *mockIEsClientMockRecorder) Query(chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*mockIEsClient)(nil).Query), chain)
}

// QueryCtx mocks base method.
func (m *mockIEsClient) QueryCtx(ctx context.Context, chain QueryChain) (*v7elastic.SearchResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryCtx", ctx, chain)
	ret0, _ := ret[0].(*v7elastic.SearchResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryCtx indicates an expected call of QueryCtx.
func (mr *mockIEsClientMockRecorder) QueryCtx(ctx, chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryCtx", reflect.TypeOf((*mockIEsClient)(nil).QueryCtx), ctx, chain)
}

// Upsert mocks base method.
func (m *mockIEsClient) Upsert(chain UpsertChain) (*v7elastic.UpdateResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", chain)
	ret0, _ := ret[0].(*v7elastic.UpdateResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upsert indicates an expected call of Upsert.
func (mr *mockIEsClientMockRecorder) Upsert(chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*mockIEsClient)(nil).Upsert), chain)
}

// UpsertCtx mocks base method.
func (m *mockIEsClient) UpsertCtx(ctx context.Context, chain UpsertChain) (*v7elastic.UpdateResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertCtx", ctx, chain)
	ret0, _ := ret[0].(*v7elastic.UpdateResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertCtx indicates an expected call of UpsertCtx.
func (mr *mockIEsClientMockRecorder) UpsertCtx(ctx, chain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertCtx", reflect.TypeOf((*mockIEsClient)(nil).UpsertCtx), ctx, chain)
}
