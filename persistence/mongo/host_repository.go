package mongo

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
)

type HostRepository struct {
	collection *mongo.Collection
}

func NewHostRepository(db *mongo.Database) *HostRepository {
	return &HostRepository{collection: db.Collection("execution_hosts")}
}

type hostDocument struct {
	ID        string          `bson:"_id"`
	Name      string          `bson:"name"`
	Adapter   string          `bson:"adapter"`
	Config    json.RawMessage `bson:"config"`
	CreatedAt time.Time       `bson:"created_at"`
	UpdatedAt time.Time       `bson:"updated_at"`
}

func toDomainHost(doc *hostDocument) *domain.ExecutionHost {
	if doc == nil {
		return nil
	}
	return &domain.ExecutionHost{
		ID:        doc.ID,
		Name:      doc.Name,
		Adapter:   doc.Adapter,
		Config:    doc.Config,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}

func (r *HostRepository) Create(ctx context.Context, hostModel *domain.ExecutionHost) error {
	_, err := r.collection.InsertOne(ctx, hostDocument{
		ID:        hostModel.ID,
		Name:      hostModel.Name,
		Adapter:   hostModel.Adapter,
		Config:    hostModel.Config,
		CreatedAt: hostModel.CreatedAt,
		UpdatedAt: hostModel.UpdatedAt,
	})
	return err
}

func (r *HostRepository) GetByID(ctx context.Context, id string) (*domain.ExecutionHost, error) {
	var doc hostDocument
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrHostNotFound
		}
		return nil, err
	}
	return toDomainHost(&doc), nil
}

func (r *HostRepository) List(ctx context.Context) ([]*domain.ExecutionHost, error) {
	cursor, err := r.collection.Find(ctx, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	hosts := make([]*domain.ExecutionHost, 0)
	for cursor.Next(ctx) {
		var doc hostDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		hosts = append(hosts, toDomainHost(&doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return hosts, nil
}
