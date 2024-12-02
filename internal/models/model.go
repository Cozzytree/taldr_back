package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Shape struct {
	Id    string                 `json:"id" bson:"id"`
	Type  string                 `json:"type" bson:"type"`
	Props map[string]interface{} `json:"props" bson:"props"`
}

type ShapeProps struct {
	Shapes []Shape `json:"shapes" bson:"shapes"`
}

type Workspace struct {
	Name        string             `json:"name" bson:"name"`
	UserId      string             `json:"userId" bson:"userId"`
	Description string             `json:"description" bson:"description"`
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Shapes      []Shape            `json:"shapes" bson:"shapes"`
	Document    string             `json:"document" bson:"document"`
}
