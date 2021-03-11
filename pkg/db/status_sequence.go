package db

import "github.com/m-mizutani/golambda"

// RetrieveStatusSequence increments sequence number of scan operation in DynamoDB and returns it.
func (x *DynamoClient) RetrieveStatusSequence() (int64, error) {
	pk := "meta:sequence"
	sk := "status"

	var result dynamoMetaSequence
	var inc int64 = 1
	query := x.table.Update("pk", pk).Range("sk", sk).Add("seq", inc)

	if err := query.Value(&result); err != nil {
		return 0, golambda.WrapError(err, "Update status sequence")
	}

	return result.Seq, nil
}
