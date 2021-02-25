package models

type DBBaseRecord struct {
	PK string `dynamo:"pk,hash" json:"-"`
	SK string `dynamo:"sk,range" json:"-"`

	PK2 string `dynamo:"pk2,omitempty" index:"secondary,hash" json:"-"`
	SK2 string `dynamo:"sk2,omitempty" index:"secondary,range" json:"-"`
}
