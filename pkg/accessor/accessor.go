package accessor

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Accessor contain mongodb client and collections
type Accessor struct {
	ctx        context.Context
	client     *mongo.Client
	collection *mongo.Collection
}

// NewAccessor makes new connection to mongoDB using target URI and etc
func NewAccessor(ctx context.Context, uri, database, collection string) (*Accessor, error) {
	// connect mongodb
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	defer c.Disconnect(ctx)

	coll := c.Database(database).Collection(collection)

	return &Accessor{
		ctx:        ctx,
		client:     c,
		collection: coll,
	}, nil
}

// Disconnect close the connection form mongDB
func (acc *Accessor) Disconnect() {
	acc.client.Disconnect(acc.ctx)
}
