package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config MongoDB 配置
type Config struct {
	URI      string
	Database string
}

// Client MongoDB 客户端包装器
type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient 创建新的 MongoDB 客户端
func NewClient(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置客户端选项
	clientOptions := options.Client().ApplyURI(cfg.URI)
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMinPoolSize(10)
	clientOptions.SetMaxConnIdleTime(30 * time.Minute)

	// 连接到 MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// 测试连接
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return &Client{
		client:   client,
		database: client.Database(cfg.Database),
	}, nil
}

// GetDatabase 获取数据库实例
func (c *Client) GetDatabase() *mongo.Database {
	return c.database
}

// GetCollection 获取集合
func (c *Client) GetCollection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close 关闭连接
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

// InsertOne 插入单个文档
func (c *Client) InsertOne(ctx context.Context, collection string, document interface{}) (*mongo.InsertOneResult, error) {
	coll := c.GetCollection(collection)
	return coll.InsertOne(ctx, document)
}

// InsertMany 插入多个文档
func (c *Client) InsertMany(ctx context.Context, collection string, documents []interface{}) (*mongo.InsertManyResult, error) {
	coll := c.GetCollection(collection)
	return coll.InsertMany(ctx, documents)
}

// FindOne 查找单个文档
func (c *Client) FindOne(ctx context.Context, collection string, filter interface{}) *mongo.SingleResult {
	coll := c.GetCollection(collection)
	return coll.FindOne(ctx, filter)
}

// Find 查找多个文档
func (c *Client) Find(ctx context.Context, collection string, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	coll := c.GetCollection(collection)
	return coll.Find(ctx, filter, opts...)
}

// UpdateOne 更新单个文档
func (c *Client) UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	coll := c.GetCollection(collection)
	return coll.UpdateOne(ctx, filter, update)
}

// UpdateMany 更新多个文档
func (c *Client) UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	coll := c.GetCollection(collection)
	return coll.UpdateMany(ctx, filter, update)
}

// DeleteOne 删除单个文档
func (c *Client) DeleteOne(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	coll := c.GetCollection(collection)
	return coll.DeleteOne(ctx, filter)
}

// DeleteMany 删除多个文档
func (c *Client) DeleteMany(ctx context.Context, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	coll := c.GetCollection(collection)
	return coll.DeleteMany(ctx, filter)
}

// CountDocuments 统计文档数量
func (c *Client) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	coll := c.GetCollection(collection)
	return coll.CountDocuments(ctx, filter)
}

// Aggregate 聚合查询
func (c *Client) Aggregate(ctx context.Context, collection string, pipeline interface{}) (*mongo.Cursor, error) {
	coll := c.GetCollection(collection)
	return coll.Aggregate(ctx, pipeline)
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, collection string, model mongo.IndexModel) (string, error) {
	coll := c.GetCollection(collection)
	return coll.Indexes().CreateOne(ctx, model)
}

// DropIndex 删除索引
func (c *Client) DropIndex(ctx context.Context, collection string, name string) error {
	coll := c.GetCollection(collection)
	_, err := coll.Indexes().DropOne(ctx, name)
	return err
}

// GetClient 获取原始 MongoDB 客户端（用于高级操作）
func (c *Client) GetClient() *mongo.Client {
	return c.client
}
