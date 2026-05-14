package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/UtopikCode/quickspaces-control-plane/domain"
)

type WorkspaceRepository struct {
	collection *mongo.Collection
}

func NewWorkspaceRepository(db *mongo.Database) *WorkspaceRepository {
	return &WorkspaceRepository{collection: db.Collection("workspaces")}
}

type workspaceDocument struct {
	ID               string                  `bson:"_id"`
	Repo             string                  `bson:"repo"`
	Owner            string                  `bson:"owner"`
	Ref              string                  `bson:"ref"`
	HostID           string                  `bson:"host_id"`
	DesiredState     string                  `bson:"desired_state"`
	ActualState      string                  `bson:"actual_state"`
	ExecutionProfile domain.ExecutionProfile `bson:"execution_profile"`
	CreatedAt        time.Time               `bson:"created_at"`
	UpdatedAt        time.Time               `bson:"updated_at"`
}

func toDomainWorkspace(doc *workspaceDocument) *domain.Workspace {
	if doc == nil {
		return nil
	}
	return &domain.Workspace{
		ID:               doc.ID,
		Repo:             doc.Repo,
		Owner:            doc.Owner,
		Ref:              doc.Ref,
		HostID:           doc.HostID,
		DesiredState:     doc.DesiredState,
		ActualState:      doc.ActualState,
		ExecutionProfile: doc.ExecutionProfile,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}
}

func (r *WorkspaceRepository) Create(ctx context.Context, workspaceModel *domain.Workspace) error {
	_, err := r.collection.InsertOne(ctx, workspaceDocument{
		ID:               workspaceModel.ID,
		Repo:             workspaceModel.Repo,
		Owner:            workspaceModel.Owner,
		Ref:              workspaceModel.Ref,
		HostID:           workspaceModel.HostID,
		DesiredState:     workspaceModel.DesiredState,
		ActualState:      workspaceModel.ActualState,
		ExecutionProfile: workspaceModel.ExecutionProfile,
		CreatedAt:        workspaceModel.CreatedAt,
		UpdatedAt:        workspaceModel.UpdatedAt,
	})
	return err
}

func (r *WorkspaceRepository) GetByID(ctx context.Context, id string) (*domain.Workspace, error) {
	var doc workspaceDocument
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, err
	}
	return toDomainWorkspace(&doc), nil
}

func (r *WorkspaceRepository) List(ctx context.Context) ([]*domain.Workspace, error) {
	cursor, err := r.collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"created_at", -1}}))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	result := make([]*domain.Workspace, 0)
	for cursor.Next(ctx) {
		var doc workspaceDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		result = append(result, toDomainWorkspace(&doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *WorkspaceRepository) UpdateDesiredState(ctx context.Context, id, desiredState string, updatedAt time.Time) error {
	res, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"desired_state": desiredState, "updated_at": updatedAt}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}

func (r *WorkspaceRepository) UpdateActualState(ctx context.Context, id, actualState string, updatedAt time.Time) error {
	res, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"actual_state": actualState, "updated_at": updatedAt}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrWorkspaceNotFound
	}
	return nil
}
