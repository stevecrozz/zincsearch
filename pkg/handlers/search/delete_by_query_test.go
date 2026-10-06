package search

import (
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/pkg/ider"
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/test/utils"
)

type arg struct {
	doc    map[string]interface{}
	query  string
	params map[string]string
}

type body struct {
	is       string
	contains string
}

type success struct {
	outcome    bool
	statusCode int
	body       body
}

type failure struct {
	statusCode int
	body       body
}

type want struct {
	success success
	failure failure
}

func TestDeleteByQuery(t *testing.T) {
	tests := []struct {
		name string
		arg  arg
		want want
	}{
		{
			name: "should delete matched documents",
			arg: arg{
				doc: map[string]interface{}{
					"name": "zinc",
				},
				query: `{"query":{"match":{"name":"zinc"}},"size":10}`,
				params: map[string]string{
					"target": "TestDeleteByQuery.index",
				},
			},
			want: want{
				success: success{
					outcome:    true,
					statusCode: 200,
					body: body{
						contains: `"time_out":false,"total":1,"deleted":1,"batches":1,"version_conflicts":0,"noops":0,"failures":[],"retries":{"bulk":0,"search":0},"throttled_millis":0,"requests_per_second":-1,"throttled_until_millis":0}`,
					},
				},
			},
		},
		{
			name: "should return bad request with invalid json body",
			arg: arg{
				doc: map[string]interface{}{
					"name": "zinc",
				},
				query: `invalid { json }`,
				params: map[string]string{
					"target": "TestDeleteByQuery.index",
				},
			},
			want: want{
				failure: failure{
					statusCode: 400,
					body: body{
						is: `{"error":"invalid character 'i' looking for beginning of value"}`,
					},
				},
			},
		},
		{
			name: "should return bad request when no matching indices are found",
			arg: arg{
				doc: map[string]interface{}{
					"name": "zinc",
				},
				query: `{"query":{"match":{"name":"zinc"}},"size":10}`,
				params: map[string]string{
					"target": "noneMatchingIndex",
				},
			},
			want: want{
				failure: failure{
					statusCode: 404,
					body: body{
						contains: `"reason":"no such index [noneMatchingIndex]","root_cause"`,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			index, err := core.NewIndex("TestDeleteByQuery.index", "disk", 2)
			assert.NoError(t, err)
			assert.NoError(t, core.StoreIndex(index))
			id := ider.Generate()
			assert.NoError(t, index.CreateDocument(id, test.arg.doc, false))
			assert.NoError(t, index.Refresh())

			c, w := utils.NewGinContext()
			utils.SetGinRequestData(c, test.arg.query)
			utils.SetGinRequestParams(c, test.arg.params)
			DeleteByQuery(c)

			if test.want.success.outcome {
				assertHTTPResponse(t, w, test.want.success.statusCode, test.want.success.body)
				assertZeruResultQuery(t, index, test.arg.query)
			} else {
				assertHTTPResponse(t, w, test.want.failure.statusCode, test.want.failure.body)
			}

			assert.NoError(t, core.DeleteIndex(index.GetName()))
		})
	}
}

func assertHTTPResponse(t *testing.T, w *httptest.ResponseRecorder, statusCode int, body body) {
	assert.Equal(t, w.Code, statusCode)
	if body.is != "" {
		assert.Equal(t, w.Body.String(), body.is)
	} else {
		assert.Contains(t, w.Body.String(), body.contains)
	}
}

func assertZeruResultQuery(t *testing.T, index *core.Index, query interface{}) {
	jsonQuery, err := json.Marshal(&query)
	assert.NoError(t, err)
	search, serr := index.Search(&meta.ZincQuery{
		Query: &meta.Query{
			Match: map[string]*meta.MatchQuery{
				"_all": {
					Query: string(jsonQuery),
				},
			},
		},
		Size: 10,
	})
	assert.NoError(t, serr)
	assert.Equal(t, 0, search.Hits.Total.Value)
}

func newDeleteByQueryIndex(t *testing.T, name string, docs int) *core.Index {
	t.Helper()
	index, err := core.NewIndex(name, "disk", 2)
	require.NoError(t, err)
	require.NoError(t, core.StoreIndex(index))
	t.Cleanup(func() { _ = core.DeleteIndex(name) })
	for i := 0; i < docs; i++ {
		require.NoError(t, index.CreateDocument(ider.Generate(), map[string]interface{}{"n": i}, false))
	}
	require.NoError(t, index.Refresh())
	return index
}

func deleteByQuery(target string, query map[string]string) *httptest.ResponseRecorder {
	c, w := utils.NewGinContext()
	utils.SetGinRequestURL(c, "/"+target+"/_delete_by_query", query)
	utils.SetGinRequestData(c, `{"query":{"match_all":{}}}`)
	utils.SetGinRequestParams(c, map[string]string{"target": target})
	DeleteByQuery(c)
	return w
}

func countDocs(t *testing.T, index *core.Index) int {
	t.Helper()
	resp, err := index.Search(&meta.ZincQuery{Query: &meta.Query{MatchAll: &meta.MatchAllQuery{}}, Size: 0})
	require.NoError(t, err)
	return resp.Hits.Total.Value
}

func TestDeleteByQueryMultipleTargets(t *testing.T) {
	a := newDeleteByQueryIndex(t, "TestDeleteByQueryMulti.a", 2)
	b := newDeleteByQueryIndex(t, "TestDeleteByQueryMulti.b", 3)
	require.NoError(t, core.ZINC_INDEX_ALIAS_LIST.AddIndexesToAlias("TestDeleteByQueryMulti.alias_b", []string{b.GetName()}))
	t.Cleanup(func() { _ = core.ZINC_INDEX_ALIAS_LIST.RemoveIndexFromAllAliases(b.GetName()) })

	w := deleteByQuery("TestDeleteByQueryMulti.a,TestDeleteByQueryMulti.alias_b", map[string]string{"refresh": "true"})

	assert.Equal(t, 200, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"total":5,"deleted":5`)
	assert.Equal(t, 0, countDocs(t, a))
	assert.Equal(t, 0, countDocs(t, b))
}

func TestDeleteByQueryIgnoreUnavailable(t *testing.T) {
	a := newDeleteByQueryIndex(t, "TestDeleteByQueryIgnore.a", 1)

	w := deleteByQuery("TestDeleteByQueryIgnore.a,TestDeleteByQueryIgnore.missing", map[string]string{"ignore_unavailable": "true", "refresh": "true"})

	assert.Equal(t, 200, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"total":1,"deleted":1`)
	assert.Equal(t, 0, countDocs(t, a))
}

func TestDeleteByQueryMissingTarget(t *testing.T) {
	newDeleteByQueryIndex(t, "TestDeleteByQueryMissing.a", 1)

	w := deleteByQuery("TestDeleteByQueryMissing.a,TestDeleteByQueryMissing.missing", nil)

	assert.Equal(t, 404, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"type":"index_not_found_exception"`)
	assert.Contains(t, w.Body.String(), `"index":"TestDeleteByQueryMissing.missing"`)
}

func TestDeleteByQueryDeletesEveryMatch(t *testing.T) {
	index := newDeleteByQueryIndex(t, "TestDeleteByQueryMany.index", 25)

	w := deleteByQuery(index.GetName(), map[string]string{"refresh": "true", "conflicts": "proceed"})

	assert.Equal(t, 200, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), `"total":25,"deleted":25`)
	assert.Equal(t, 0, countDocs(t, index))
}
