package db

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

const (
	dynamoGSIName = "secondary"
)

type dynamoRecord struct {
	PK string `dynamo:"pk,hash"`
	SK string `dynamo:"sk,range"`

	PK2 string `dynamo:"pk2,omitempty" index:"secondary,hash"`
	SK2 string `dynamo:"sk2,omitempty" index:"secondary,range"`

	Docs interface{} `dynamo:"docs"`
}

// DynamoClient is implementation of interfaces.DBClient to use Amazon DynamoDB
type DynamoClient struct {
	tableName string
	table     dynamo.Table
	local     bool
}

// TableName is to identify name of table created for local test
func (x *DynamoClient) TableName() string { return x.tableName }

// NewDynamoClient creates DynamoClient
func NewDynamoClient(region, tableName string) (*DynamoClient, error) {
	ssn, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	table := dynamo.New(ssn).Table(tableName)
	return &DynamoClient{
		table: table,
	}, nil
}

// NewDynamoClientLocal configures DynamoClient with local endpoint and create a table for test and return the client.
func NewDynamoClientLocal(region, tableName string) (*DynamoClient, error) {
	// Set port number
	port := 8000
	if v, ok := os.LookupEnv("DYNAMO_LOCAL_PORT"); ok {
		localPort, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			panic("DYNAMO_LOCAL_PORT can not be parsed: " + v)
		}
		if 65535 < localPort {
			panic("DYNAMO_LOCAL_PORT has invalid port number")
		}
		port = int(localPort)
	}

	// Add table name suffix to isolate from other test
	tableName += "-" + uuid.New().String()

	// Dummy credential
	os.Setenv("AWS_ACCESS_KEY_ID", "x")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "x")
	ssn, err := session.NewSession(&aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(fmt.Sprintf("http://localhost:%d", port)),
		// Credentials: credentials.NewStaticCredentials("dummy_key", "dummy_secret", "dummy_token"),
	})
	if err != nil {
		return nil, err
	}

	db := dynamo.New(ssn)
	if err := db.CreateTable(tableName, dynamoRecord{}).OnDemand(true).Run(); err != nil {
		return nil, golambda.WrapError(err, "Creating local DynamoDB table")
	}

	table := dynamo.New(ssn).Table(tableName)
	return &DynamoClient{
		local:     true,
		table:     table,
		tableName: tableName,
	}, nil
}

// Close deletes table if table is in local DynamoDB.
func (x *DynamoClient) Close() error {
	if x.local {
		if err := x.table.DeleteTable().Run(); err != nil {
			return err
		}
	}
	return nil
}

// Unmarshal copy record values to v via encoding and decoding as JSON.
func (x *dynamoRecord) Unmarshal(v interface{}) error {
	raw, err := json.Marshal(x.Docs)
	if err != nil {
		return golambda.WrapError(err, "json.Marshal").With("x", x)
	}

	if err := json.Unmarshal(raw, v); err != nil {
		return golambda.WrapError(err, "json.Unmarshal").With("x", x).With("raw", string(raw))
	}

	return nil
}
