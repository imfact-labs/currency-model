package mongodbstorage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"

	"github.com/imfact-labs/imfact-currency/digest/util"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"
)

var errorDuplicateKey = 11000

type (
	getRecordCallback  func(*mongo.SingleResult) error
	getRecordsCallback func(*mongo.Cursor) (bool, error)
)

type Client struct {
	uri            string
	client         *mongo.Client
	db             *mongo.Database
	connectTimeout time.Duration
	execTimeout    time.Duration
}

func NewClient(uri string, connectTimeout, execTimeout time.Duration) (*Client, error) {
	var cs connstring.ConnString
	if c, err := checkURI(uri); err != nil {
		return nil, err
	} else {
		cs = c
	}

	clientOpts := options.Client().ApplyURI(uri)
	if err := clientOpts.Validate(); err != nil {
		return nil, err
	}

	var client *mongo.Client
	{
		if c, err := mongo.Connect(clientOpts); err != nil {
			return nil, errors.Wrap(err, "connect timeout")
		} else {
			client = c
		}
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		defer cancel()

		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			return nil, errors.Wrap(err, "ping timeout")
		}
	}

	return &Client{
		uri:            uri,
		client:         client,
		db:             client.Database(cs.Database),
		connectTimeout: connectTimeout,
		execTimeout:    execTimeout,
	}, nil
}

func (cl *Client) MongoClient() *mongo.Client {
	return cl.client
}

func (cl *Client) Collection(col string) *mongo.Collection {
	return cl.db.Collection(col)
}

func (cl *Client) Collections() ([]string, error) {
	return cl.db.ListCollectionNames(context.TODO(), bson.M{})
}

func (cl *Client) Database(db string) *mongo.Database {
	return cl.client.Database(db)
}

func (cl *Client) Databases(filter interface{}) ([]string, error) {
	r, err := cl.client.ListDatabases(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	ds := r.Databases
	if len(ds) < 1 {
		return nil, nil
	}

	l := make([]string, len(ds))
	for i := range ds {
		l[i] = ds[i].Name
	}

	return l, nil
}

func (cl *Client) Find(
	ctx context.Context,
	col string,
	query interface{},
	callback getRecordsCallback,
	opts ...options.Lister[options.FindOptions],
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	var cursor *mongo.Cursor
	if c, err := cl.db.Collection(col).Find(ctx, query, opts...); err != nil {
		return err
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
			defer cancel()

			_ = c.Close(ctx)
		}()

		cursor = c
	}

	next := func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
		defer cancel()

		return cursor.Next(ctx)
	}

	for next() {
		if keep, e := callback(cursor); e != nil {
			return e
		} else if !keep {
			break
		}
	}

	if err := cursor.Err(); err != nil {
		return errors.Errorf("cursor error: %v", err)
	}

	return nil
}

func (cl *Client) GetByID(
	col string,
	id interface{},
	callback getRecordCallback,
	opts ...options.Lister[options.FindOneOptions],
) error {
	res, err := cl.getByFilter(col, util.NewBSONFilter("_id", id).D(), opts...)
	if err != nil {
		return err
	}

	if callback == nil {
		return nil
	}

	return callback(res)
}

func (cl *Client) GetByFilter(
	col string,
	filter bson.D,
	callback getRecordCallback,
	opts ...options.Lister[options.FindOneOptions],
) error {
	res, err := cl.getByFilter(col, filter, opts...)
	if err != nil {
		return err
	}

	if callback == nil {
		return nil
	}

	return callback(res)
}

func (cl *Client) getByFilter(col string, filter bson.D, opts ...options.Lister[options.FindOneOptions]) (*mongo.SingleResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	res := cl.db.Collection(col).FindOne(ctx, filter, opts...)
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, err
		}

		return nil, err
	}

	return res, nil
}

func (cl *Client) Aggregate(
	ctx context.Context,
	col string,
	pipeline interface{}, // mongo.Pipeline 또는 []bson.D
	callback getRecordsCallback,
	opts ...options.Lister[options.AggregateOptions],
) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	var cursor *mongo.Cursor
	c, err := cl.db.Collection(col).Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return err
	}

	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
		defer cancel()
		_ = c.Close(closeCtx)
	}()
	cursor = c

	next := func() bool {
		nextCtx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
		defer cancel()
		return cursor.Next(nextCtx)
	}

	for next() {
		if keep, e := callback(cursor); e != nil {
			return e
		} else if !keep {
			break
		}
	}

	if err := cursor.Err(); err != nil {
		return errors.Errorf("cursor error: %v", err)
	}

	return nil
}

func (cl *Client) Add(col string, doc Doc) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	res, err := cl.db.Collection(col).InsertOne(ctx, doc)
	if err != nil {
		if isDuplicatedError(err) {
			return nil, err
		}

		return nil, err
	}

	return res.InsertedID, nil
}

func (cl *Client) Set(col string, doc Doc) (interface{}, error) {
	if doc.ID() == nil {
		return cl.Add(col, doc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	opts := options.Replace().SetUpsert(true) // 없으면 넣고, 있으면 통째로 교체
	filter := util.NewBSONFilter("_id", doc.ID()).D()

	res, err := cl.db.Collection(col).ReplaceOne(ctx, filter, doc, opts)
	if err != nil {
		return nil, err
	}

	if res.UpsertedCount > 0 && res.UpsertedID != nil {
		return res.UpsertedID, nil
	}

	return doc.ID(), nil
}

func (cl *Client) AddRaw(col string, raw bson.Raw) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	res, err := cl.db.Collection(col).InsertOne(ctx, raw)
	if err != nil {
		if isDuplicatedError(err) {
			return nil, err
		}

		return nil, err
	}

	return res.InsertedID, nil
}

func (cl *Client) Bulk(ctx context.Context, col string, models []mongo.WriteModel, order bool) error {
	opts := options.BulkWrite().SetOrdered(order)
	if _, err := writeBulkModels(ctx, cl, col, models, defaultLimitWriteModels, opts); err != nil {
		return err
	} else {
		return nil
	}
}

func (cl *Client) Count(ctx context.Context, col string, filter interface{}, opts ...options.Lister[options.CountOptions]) (int64, error) {
	count, err := cl.db.Collection(col).CountDocuments(ctx, filter, opts...)

	return count, err
}

func (cl *Client) Delete(col string, filter bson.D, opts ...options.Lister[options.DeleteManyOptions]) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	return cl.db.Collection(col).DeleteMany(ctx, filter, opts...)
}

func (cl *Client) Exists(col string, filter bson.D) (bool, error) {
	count, err := cl.Count(context.Background(), col, filter, options.Count().SetLimit(1))

	return count > 0, err
}

func (cl *Client) WithSession(
	ctx context.Context,
	callback func(context.Context, func(string /* collection */) *mongo.Collection) (interface{}, error),
) (interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, cl.execTimeout)
	defer cancel()

	txnDefaults := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority()).
		SetReadPreference(readpref.Primary())

	sessOpts := options.Session().
		SetCausalConsistency(true).
		SetDefaultTransactionOptions(txnDefaults)
	sess, err := cl.client.StartSession(sessOpts)
	if err != nil {
		return nil, err
	}

	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
		defer cancel()
		sess.EndSession(closeCtx)
	}()

	result, err := sess.WithTransaction(
		ctx,
		func(txnCtx context.Context) (interface{}, error) {
			return callback(txnCtx, cl.Collection)
		},
		txnDefaults,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (cl *Client) DropDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	return cl.db.Drop(ctx)
}

func (cl *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), cl.execTimeout)
	defer cancel()

	return cl.client.Disconnect(ctx)
}

func (cl *Client) Raw() *mongo.Client {
	return cl.client
}

func (cl *Client) CopyCollection(source *Client, fromCol, toCol string) error {
	var limit = 100
	var models []mongo.WriteModel
	err := source.Find(context.Background(), fromCol, bson.D{}, func(cursor *mongo.Cursor) (bool, error) {
		if len(models) == limit {
			if err := cl.Bulk(context.Background(), toCol, models, false); err != nil {
				return false, err
			} else {
				models = nil
			}
		}

		raw := util.CopyBytes(cursor.Current)
		models = append(models, mongo.NewInsertOneModel().SetDocument(bson.Raw(raw)))

		return true, nil
	})
	if err != nil {
		return err
	}

	if len(models) < 1 {
		return nil
	}

	return cl.Bulk(context.Background(), toCol, models, false)
}

func (cl *Client) New(db string) (*Client, error) {
	var d *mongo.Database
	if len(db) < 1 {
		d = cl.db
	} else {
		d = cl.client.Database(db)
	}

	return &Client{
		uri:            cl.uri,
		client:         cl.client,
		db:             d,
		connectTimeout: cl.connectTimeout,
		execTimeout:    cl.execTimeout,
	}, nil
}

func isDuplicatedError(err error) bool {
	switch t := err.(type) {
	case mongo.WriteException:
		for i := range t.WriteErrors {
			if t.WriteErrors[i].Code == errorDuplicateKey {
				return true
			}
		}

		return false
	case mongo.BulkWriteException:
		for i := range t.WriteErrors {
			if t.WriteErrors[i].WriteError.Code == errorDuplicateKey {
				return true
			}
		}

		return false
	default:
		return false
	}
}
