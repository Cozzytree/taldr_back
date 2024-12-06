package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Cozzytree/taldrBack/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() string
	SaveShapetoDb(ctx context.Context,
		workSpaceId string,
		userId string,
		shapes []models.Shape) error
	SaveDocsToDb(
		ctx context.Context,
		workSpaceId string,
		userId string,
		document string) error
	NewWorkSpace(ctx context.Context, workspace models.Workspace) error
	UserWorkspaces(ctx context.Context, userId string, includeShapes bool) ([]models.Workspace, error)
	GetWorkspaceData(ctx context.Context, workspaceId primitive.ObjectID) (*models.Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceId primitive.ObjectID, userId string) error
	UpdateNameDescription(ctx context.Context,
		workspaceId primitive.ObjectID,
		userId string,
		description string,
		name string) error
	UpdateDescription(
		ctx context.Context,
		workspaceId primitive.ObjectID,
		description string,
		userId string) error
	UpdateName(ctx context.Context, workspaceId primitive.ObjectID, name string, userId string) error
}

type service struct {
	db *mongo.Client
}

func New() Service {
	// uri := os.Getenv("MONGODB_URI")
	uri := "mongodb://root:secret@localhost:27017/?authMechanism=SCRAM-SHA-1&retryWrites=true&w=majority"
	// uri := "mongodb://root:secret@taldr:27017/?authMechanism=SCRAM-SHA-1&retryWrites=true&w=majority"
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Fatal("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
	}
	client, err := mongo.
		Connect(context.TODO(), options.Client().
			ApplyURI(uri))

	if err != nil {
		panic(err)
	}

	fmt.Println("connected to db")

	return &service{
		db: client,
	}
}

func (s *service) Health() string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	return "Healthy"
}

func (s *service) SaveShapetoDb(
	ctx context.Context,
	workSpaceId string,
	userId string,
	shapes []models.Shape) error {

	id, err := primitive.ObjectIDFromHex(workSpaceId)
	if err != nil {
		return err
	}

	conn := s.db.Database("taldr").Collection("workspaces")

	filter := bson.M{"userId": userId, "_id": id}
	res := conn.FindOne(ctx, filter)
	if res.Err() != nil {
		return res.Err()
	}

	res = conn.FindOneAndUpdate(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{"shapes": shapes}},
	)

	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (s *service) SaveDocsToDb(
	ctx context.Context,
	workSpaceId string,
	userId string,
	document string) error {
	id, err := primitive.ObjectIDFromHex(workSpaceId)
	if err != nil {
		return err
	}

	conn := s.db.Database("taldr").Collection("workspaces")

	filter := bson.M{"userId": userId, "_id": id}
	res := conn.FindOne(ctx, filter)
	if res.Err() != nil {
		return res.Err()
	}

	res = conn.FindOneAndUpdate(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{"document": document}},
	)

	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (s *service) NewWorkSpace(ctx context.Context,
	workspace models.Workspace) error {
	conn := s.db.Database("taldr").Collection("workspaces")

	workspace.ID = primitive.NewObjectID()
	_, err := conn.InsertOne(ctx, workspace)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UserWorkspaces(ctx context.Context,
	userId string, includeShapes bool) ([]models.Workspace, error) {

	var project bson.M
	if includeShapes {
		project = bson.M{"shapes": 1, "name": 1, "description": 1, "userId": 1}
	} else {
		project = bson.M{"name": 1, "description": 1, "userId": 1}
	}

	opts := options.Find().SetProjection(project)

	conn, err := s.db.Database("taldr").
		Collection("workspaces").
		Find(ctx, bson.M{"userId": userId}, opts)

	if err != nil {
		return nil, err
	}

	var works []models.Workspace
	err = conn.All(ctx, &works)
	if err != nil {
		return nil, err
	}

	return works, nil
}

func (s *service) GetWorkspaceData(
	ctx context.Context,
	workspaceId primitive.ObjectID) (*models.Workspace, error) {
	// FindOne will return an error if no document is found, or other database issues.
	res := s.db.Database("taldr").
		Collection("workspaces").
		FindOne(ctx, bson.M{"_id": workspaceId})

	// Check for errors in fetching the document.
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			// If no document is found, return a custom error or nil.
			return nil, fmt.Errorf("workspace not found")
		}
		return nil, fmt.Errorf("error finding workspace: %v", res.Err())
	}

	// Decode the result into the Workspace struct.
	var data models.Workspace
	err := res.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("error decoding workspace data: %v", err)
	}

	// Return the decoded workspace data.
	return &data, nil
}

func (s *service) DeleteWorkspace(
	ctx context.Context,
	workspaceId primitive.ObjectID,
	userId string) error {
	conn := s.db.Database("taldr").Collection("workspaces")

	res := conn.FindOne(ctx, bson.M{"_id": workspaceId})
	if res.Err() != nil {
		return res.Err()
	}

	var isowner models.Workspace

	err := res.Decode(&isowner)
	if err != nil {
		return errors.New("internal server error")
	}

	if isowner.UserId != userId {
		return errors.New("umauthorized")
	}

	res = conn.FindOneAndDelete(ctx, bson.M{"_id": workspaceId})
	if res.Err() != nil {
		return errors.New("err : %v" + res.Err().Error())
	}
	return nil
}

func (s *service) UpdateNameDescription(ctx context.Context,
	workspaceId primitive.ObjectID,
	userId string,
	description string,
	name string) error {

	conn := s.db.Database("taldr").Collection("workspaces")
	res := conn.FindOne(ctx, bson.M{"_id": workspaceId})
	if res.Err() != nil {
		return res.Err()
	}

	var isowner models.Workspace

	err := res.Decode(&isowner)
	if err != nil {
		return errors.New("internal server error")
	}

	if isowner.UserId != userId {
		return errors.New("umauthorized")
	}

	update := bson.M{"$set": bson.M{"description": description, "name": name}}

	res = conn.FindOneAndUpdate(ctx, bson.M{"_id": workspaceId}, update)
	if res.Err() != nil {
		return errors.New("err : %v" + res.Err().Error())
	}
	return nil
}

func (s *service) UpdateName(
	ctx context.Context,
	workspaceId primitive.ObjectID,
	name string,
	userId string) error {
	conn := s.db.Database("taldr").Collection("workspaces")

	exist := conn.FindOne(ctx, bson.M{"_id": workspaceId})
	if exist.Err() != nil {
		return exist.Err()
	}

	var workspace models.Workspace
	if err := exist.Decode(&workspace); err != nil {
		return errors.New("internal server error")
	}

	if workspace.UserId != userId {
		return errors.New("unauthorized")
	}

	updateRes := conn.FindOneAndUpdate(ctx, bson.M{"_id": workspaceId}, bson.M{"$set": bson.M{"name": name}})
	if updateRes.Err() != nil {
		return updateRes.Err()
	}

	return nil
}

func (s *service) UpdateDescription(
	ctx context.Context,
	workspaceId primitive.ObjectID,
	description string,
	userId string) error {
	conn := s.db.Database("taldr").Collection("workspaces")

	exist := conn.FindOne(ctx, bson.M{"_id": workspaceId})
	if exist.Err() != nil {
		return exist.Err()
	}

	var workspace models.Workspace
	if err := exist.Decode(&workspace); err != nil {
		return errors.New("internal server error")
	}

	if workspace.UserId != userId {
		return errors.New("unauthorized")
	}

	updateRes := conn.
		FindOneAndUpdate(ctx, bson.M{"_id": workspaceId}, bson.M{"$set": bson.M{"description": description}})
	if updateRes.Err() != nil {
		return updateRes.Err()
	}

	return nil
}
