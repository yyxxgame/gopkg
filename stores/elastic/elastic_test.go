package elastic

import (
	v7elastic "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestQuery(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := NewMockIEsClient(mockCtrl)

	expectedResults := &v7elastic.SearchResult{
		Hits: &v7elastic.SearchHits{
			TotalHits: &v7elastic.TotalHits{Value: 1},
			Hits: []*v7elastic.SearchHit{
				{
					Index:  "test-index",
					Source: []byte(`{"title": "Test Document"}`),
				},
			},
		},
	}

	mockClient.EXPECT().Query(gomock.Any()).Return(expectedResults, nil)

	result, err := mockClient.Query(func(srv *v7elastic.SearchService) *v7elastic.SearchService {
		return srv
	})
	assert.Nil(t, err)

	assert.Equal(t, expectedResults.TotalHits(), result.TotalHits())
}

func TestInsert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := NewMockIEsClient(mockCtrl)

	expectedResults := &v7elastic.IndexResponse{
		Index: "test-index",
	}

	mockClient.EXPECT().Insert(gomock.Any()).Return(expectedResults, nil)

	result, err := mockClient.Insert(func(srv *v7elastic.IndexService) *v7elastic.IndexService {
		return srv
	})
	assert.Nil(t, err)

	assert.Equal(t, expectedResults.Index, result.Index)
}

func TestUpsert(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := NewMockIEsClient(mockCtrl)

	expectedResults := &v7elastic.UpdateResponse{
		Index: "test-index",
	}

	mockClient.EXPECT().Upsert(gomock.Any()).Return(expectedResults, nil)

	result, err := mockClient.Upsert(func(srv *v7elastic.UpdateService) *v7elastic.UpdateService {
		return srv
	})
	assert.Nil(t, err)

	assert.Equal(t, expectedResults.Index, result.Index)
}
