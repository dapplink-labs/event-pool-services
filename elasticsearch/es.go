package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"

	"github.com/multimarket-labs/event-pod-services/common/retry"
	"github.com/multimarket-labs/event-pod-services/config"
)

type ESClient struct {
	client *elasticsearch.Client
	config config.ESConfig
}

// NewESClient 创建新的Elasticsearch客户端
func NewESClient(ctx context.Context, esConfig config.ESConfig) (*ESClient, error) {
	if !esConfig.Enable {
		return nil, nil // 如果未启用，返回nil
	}

	cfg := elasticsearch.Config{
		Addresses: esConfig.Addresses,
	}

	// 配置认证
	if esConfig.Username != "" && esConfig.Password != "" {
		cfg.Username = esConfig.Username
		cfg.Password = esConfig.Password
	} else if esConfig.APIKey != "" {
		cfg.APIKey = esConfig.APIKey
	} else if esConfig.ServiceToken != "" {
		cfg.ServiceToken = esConfig.ServiceToken
	} else if esConfig.CloudID != "" {
		cfg.CloudID = esConfig.CloudID
	}

	// 如果需要自定义Transport，可以在这里配置
	// 默认情况下，Elasticsearch客户端会自动处理HTTP/HTTPS

	// 使用重试策略连接
	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	client, err := retry.Do[*elasticsearch.Client](ctx, 10, retryStrategy, func() (*elasticsearch.Client, error) {
		es, err := elasticsearch.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
		}

		// 测试连接
		res, err := es.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			return nil, fmt.Errorf("elasticsearch info error: %s", res.String())
		}

		return es, nil
	})

	if err != nil {
		return nil, err
	}

	return &ESClient{
		client: client,
		config: esConfig,
	}, nil
}

// Index 索引文档
func (es *ESClient) Index(ctx context.Context, index string, documentID string, body interface{}) error {
	if es == nil || es.client == nil {
		return errors.New("elasticsearch client is not initialized")
	}

	var buf strings.Builder
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("failed to encode document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: documentID,
		Body:       strings.NewReader(buf.String()),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index error: %s", res.String())
	}

	return nil
}

// Get 获取文档
func (es *ESClient) Get(ctx context.Context, index string, documentID string) (map[string]interface{}, error) {
	if es == nil || es.client == nil {
		return nil, errors.New("elasticsearch client is not initialized")
	}

	req := esapi.GetRequest{
		Index:      index,
		DocumentID: documentID,
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, errors.New("document not found")
		}
		return nil, fmt.Errorf("get error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// Search 搜索文档
func (es *ESClient) Search(ctx context.Context, index string, query map[string]interface{}) (map[string]interface{}, error) {
	if es == nil || es.client == nil {
		return nil, errors.New("elasticsearch client is not initialized")
	}

	var buf strings.Builder
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(buf.String()),
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// Delete 删除文档
func (es *ESClient) Delete(ctx context.Context, index string, documentID string) error {
	if es == nil || es.client == nil {
		return errors.New("elasticsearch client is not initialized")
	}

	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: documentID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil // 文档不存在，认为删除成功
		}
		return fmt.Errorf("delete error: %s", res.String())
	}

	return nil
}

// CreateIndex 创建索引
func (es *ESClient) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	if es == nil || es.client == nil {
		return errors.New("elasticsearch client is not initialized")
	}

	var buf strings.Builder
	if mapping != nil {
		if err := json.NewEncoder(&buf).Encode(mapping); err != nil {
			return fmt.Errorf("failed to encode mapping: %w", err)
		}
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  strings.NewReader(buf.String()),
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.String())
	}

	return nil
}

// DeleteIndex 删除索引
func (es *ESClient) DeleteIndex(ctx context.Context, index string) error {
	if es == nil || es.client == nil {
		return errors.New("elasticsearch client is not initialized")
	}

	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil // 索引不存在，认为删除成功
		}
		return fmt.Errorf("delete index error: %s", res.String())
	}

	return nil
}

// Close 关闭客户端连接
func (es *ESClient) Close() error {
	// Elasticsearch客户端不需要显式关闭
	return nil
}

// Client 返回底层客户端（用于高级操作）
func (es *ESClient) Client() *elasticsearch.Client {
	return es.client
}
