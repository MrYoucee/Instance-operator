package models

type Server struct {
	Name        string `json:"name" bson:"name,omitempty"`
	Description string `json:"description" bson:"description,omitempty"`
	Datacenter  string `json:"datacenter" bson:"datacenter,omitempty"`
}
