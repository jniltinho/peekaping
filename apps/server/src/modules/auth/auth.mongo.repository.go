package auth

import (
	"context"
	"errors"
	"peekaping/src/config"
	"peekaping/src/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoModel struct {
	ID             primitive.ObjectID `bson:"_id"`
	Email          string             `bson:"email"`
	Password       string             `bson:"password"`
	Active         bool               `bson:"active"`
	TwoFASecret    string             `bson:"twofa_secret"`
	TwoFAStatus    bool               `bson:"twofa_status"`
	TwoFALastToken string             `bson:"twofa_last_token"`
	CreatedAt      time.Time          `bson:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt"`
}

func toDomainModel(mm *mongoModel) *Model {
	return &Model{
		ID:             mm.ID.Hex(),
		Email:          mm.Email,
		Password:       mm.Password,
		Active:         mm.Active,
		TwoFASecret:    mm.TwoFASecret,
		TwoFAStatus:    mm.TwoFAStatus,
		TwoFALastToken: mm.TwoFALastToken,
		CreatedAt:      mm.CreatedAt,
		UpdatedAt:      mm.UpdatedAt,
	}
}

type RepositoryImpl struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, cfg *config.Config) Repository {
	db := client.Database(cfg.DBName)
	collection := db.Collection("users")
	return &RepositoryImpl{client, db, collection}
}

func (r *RepositoryImpl) Create(ctx context.Context, user *Model) (*Model, error) {
	mm := &mongoModel{
		ID:             primitive.NewObjectID(),
		Email:          user.Email,
		Password:       user.Password,
		Active:         user.Active,
		TwoFASecret:    user.TwoFASecret,
		TwoFAStatus:    user.TwoFAStatus,
		TwoFALastToken: user.TwoFALastToken,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, mm)
	if err != nil {
		return nil, err
	}

	return toDomainModel(mm), nil
}

func (r *RepositoryImpl) FindByEmail(ctx context.Context, email string) (*Model, error) {
	var admin mongoModel
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&admin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainModel(&admin), nil
}

func (r *RepositoryImpl) FindByID(ctx context.Context, id string) (*Model, error) {
	var entity mongoModel

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	err = r.collection.FindOne(ctx, filter).Decode(&entity)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainModel(&entity), nil
}

func (r *RepositoryImpl) Update(ctx context.Context, id string, entity *UpdateModel) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	set, err := utils.ToBsonSet(entity)
	if err != nil {
		return err
	}

	if len(set) == 0 {
		return errors.New("nothing to update")
	}

	// Always set updatedAt to current time
	set["updatedAt"] = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": set}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *RepositoryImpl) FindAllCount(ctx context.Context) (int64, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	return count, err
}
