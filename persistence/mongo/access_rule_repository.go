package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/UtopikCode/quickspaces-control-plane/internal/application/auth"
)

type AccessRuleRepository struct {
	collection *mongo.Collection
}

func NewAccessRuleRepository(db *mongo.Database) *AccessRuleRepository {
	return &AccessRuleRepository{collection: db.Collection("access_rules")}
}

type accessRuleDocument struct {
	SubjectType string    `bson:"subject_type"`
	SubjectID   string    `bson:"subject_id"`
	Role        string    `bson:"role"`
	CreatedAt   time.Time `bson:"created_at"`
}

func toDomainAccessRule(doc *accessRuleDocument) *auth.AccessRule {
	if doc == nil {
		return nil
	}
	return &auth.AccessRule{
		SubjectType: doc.SubjectType,
		SubjectID:   doc.SubjectID,
		Role:        doc.Role,
	}
}

func (r *AccessRuleRepository) List(ctx context.Context) ([]*auth.AccessRule, error) {
	cursor, err := r.collection.Find(ctx, bson.D{{}})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	rules := make([]*auth.AccessRule, 0)
	for cursor.Next(ctx) {
		var doc accessRuleDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		rules = append(rules, toDomainAccessRule(&doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *AccessRuleRepository) Upsert(ctx context.Context, subjectType, subjectID, role string) error {
	filter := bson.M{"subject_type": subjectType, "subject_id": subjectID}
	update := bson.M{
		"$set":         bson.M{"role": role},
		"$setOnInsert": bson.M{"created_at": time.Now().UTC()},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *AccessRuleRepository) Delete(ctx context.Context, subjectType, subjectID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"subject_type": subjectType, "subject_id": subjectID})
	return err
}
