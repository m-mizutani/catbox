package controller

import "github.com/m-mizutani/catbox/pkg/interfaces"

// DB returns DBClient. Create a new DBClient by NewDBClient if dbClient is still nil
func (x *Controller) DB() interfaces.DBClient {
	if x.dbClient == nil {
		client, err := x.adaptors.NewDBClient(x.Config.AwsRegion, x.Config.TableName)
		if err != nil {
			panic("Failed to NewDBClient: " + err.Error())
		}
		x.dbClient = client
	}

	return x.dbClient
}
