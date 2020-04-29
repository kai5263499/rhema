package interfaces

import (
	"context"

	"github.com/olivere/elastic/v7"
)

type ElasticSearch interface {
	Index() *elastic.IndexService
	IndexExists(indices ...string) *elastic.IndicesExistsService
	CreateIndex(name string) *elastic.IndicesCreateService
	Get() *elastic.GetService
	Update() *elastic.UpdateService
	Search(indices ...string) *elastic.SearchService
}

type IndexService interface {
	Index(index string) IndexService
	Type(typ string) IndexService
	BodyJson(body interface{}) IndexService
	Do(ctx context.Context) (*elastic.IndexResponse, error)
}
