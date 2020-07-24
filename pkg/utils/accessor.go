/**
 * Copyright 2020-2020 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	timeout    = 10 * time.Second
	timeFormat = "2006-01-02T15:04:05"
)

// Accessor contain mongodb client and collections
type Accessor struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// Record corresponds to one record in the log.
type Record struct {
	File    string   `bson:"file"`
	Level   string   `bson:"level"`
	Line    int      `bson:"line"`
	Message string   `bson:"message"`
	NID     string   `bson:"nid"`
	Param   bson.Raw `bson:"param"`
	Time    string   `bson:"time"`
	TimeNtv time.Time
}

// NewAccessor makes new connection to mongoDB using target URI and etc
func NewAccessor(uri, database, collection string) (*Accessor, error) {
	// make context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// connect mongodb
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err = client.Connect(ctx); err != nil {
		return nil, err
	}

	coll := client.Database(database).Collection(collection)

	return &Accessor{
		client:     client,
		collection: coll,
	}, nil
}

// GetEarliestTime gets the timestamp of the earliest record in the DB
func (acc *Accessor) GetEarliestTime() (*time.Time, error) {
	var result Record
	option := options.FindOne().SetSort(bson.M{"time": 1})
	err := acc.collection.FindOne(context.Background(), bson.D{}, option).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// drop timezone data
	result.TimeNtv, err = time.Parse(timeFormat, result.Time[0:19])

	if err != nil {
		return nil, err
	}
	return &result.TimeNtv, nil
}

// GetLastTime gets the timestamp of the last record in the DB
func (acc *Accessor) GetLastTime() (*time.Time, error) {
	var result Record
	option := options.FindOne().SetSort(bson.M{"time": -1})
	err := acc.collection.FindOne(context.Background(), bson.D{}, option).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// drop timezone data
	result.TimeNtv, err = time.Parse(timeFormat, result.Time[0:19])

	if err != nil {
		return nil, err
	}
	return &result.TimeNtv, nil
}

// GetByTime gets records for the specified time
func (acc *Accessor) GetByTime(t *time.Time) ([]Record, error) {
	option := options.Find().SetSort(bson.M{"time": 1})
	filter := bson.M{
		"time": bson.M{"$regex": "^" + t.Format(timeFormat)},
	}
	cur, err := acc.collection.Find(context.Background(), filter, option)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	results := make([]Record, 0)
	for cur.Next(context.Background()) {
		var result Record
		if err = cur.Decode(&result); err != nil {
			return nil, err
		}
		// drop timezone data
		result.TimeNtv, err = time.Parse(timeFormat, result.Time[0:19])

		results = append(results, result)
	}
	return results, nil
}

// GetByTimeMessage gets records having specified time and message
func (acc *Accessor) GetByTimeMessage(t *time.Time, message string) ([]Record, error) {
	option := options.Find().SetSort(bson.M{"time": 1})
	filter := bson.M{
		"message": message,
		"time":    bson.M{"$regex": "^" + t.Format(timeFormat)},
	}
	cur, err := acc.collection.Find(context.Background(), filter, option)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	results := make([]Record, 0)
	for cur.Next(context.Background()) {
		var result Record
		if err = cur.Decode(&result); err != nil {
			return nil, err
		}
		// drop timezone data
		result.TimeNtv, err = time.Parse(timeFormat, result.Time[0:19])

		results = append(results, result)
	}
	return results, nil
}

// Disconnect close the connection form mongDB
func (acc *Accessor) Disconnect() {
	// make context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	acc.client.Disconnect(ctx)
}
