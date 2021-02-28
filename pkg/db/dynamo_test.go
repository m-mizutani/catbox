package db_test

import (
	"testing"

	"github.com/m-mizutani/catbox/pkg/db"
)

func newTestTable(t *testing.T) *db.DynamoClient {
	tableName := "dynamodb-test"

	client, err := db.NewDynamoClientLocal("ap-northeast-1", tableName)
	if err != nil {
		panic("Failed to use local DynamoDB: " + err.Error())
	}

	t.Log("Created table name: ", client.TableName())
	return client
}

func deleteTestTable(t *testing.T, client *db.DynamoClient) {
	if t.Failed() {
		return // Failed test table is not deleted
	}

	if err := client.Close(); err != nil {
		panic("Failed to delete test table: " + err.Error())
	}
}
