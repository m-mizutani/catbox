package usecase_test

/*
func TestUpdateRepoVulnStatus(t *testing.T) {
	baseImg := model.TaggedImage{
		Registry: "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com",
		Repo:     "strix",
		Tag:      "latest",
	}
	testData := []*model.RepoVulnStatus{
		{
			TaggedImage:        baseImg,
			ScanSequence: 8,
			VulnID:       "CVE-2020-27350",
			VulnTitle:    "",
			UpdatedAt:    1000,
			Status:       model.VulnStatusNew,
			ScanType:     model.ScanTypeTrivy,
			Description:  "",
			PkgType:      "ubuntu",
			PkgSource:    "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com/strix:latest (ubuntu 18.04)",
			PkgName:      "apt",
		},
		{
			TaggedImage:        baseImg,
			ScanSequence: 9,
			VulnID:       "CVE-2020-27350",
			VulnTitle:    "",
			UpdatedAt:    1000,
			Status:       model.VulnStatusMitigated,
			ScanType:     model.ScanTypeTrivy,
			Description:  "",
			PkgType:      "ubuntu",
			PkgSource:    "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com/strix:latest (ubuntu 18.04)",
			PkgName:      "apt",
		},

		{
			TaggedImage:        baseImg,
			ScanSequence: 9,
			VulnID:       "CVE-2020-0001",
			VulnTitle:    "Should Fixed Vuln",
			UpdatedAt:    1000,
			Status:       model.VulnStatusNew,
			ScanType:     model.ScanTypeTrivy,
			Description:  "",
			PkgType:      "ubuntu",
			PkgSource:    "111111111111.dkr.ecr.ap-northeast-1.amazonaws.com/strix:latest (ubuntu 18.04)",
			PkgName:      "xyz",
		},
	}

	t.Run("Update multiple status", func(t *testing.T) {
		ctrl, mock := newControllerForRepoVulnStatusTest(t)
		require.NoError(t, mock.dbClient.PutRepoVulnStatusBatch(testData))

		usecase.UpdateRepoVulnStatuses(ctrl, []*model.RepoVulnStatus{
			{},
		})
	})
}

func newControllerForRepoVulnStatusTest(t *testing.T) (*controller.Controller, *mockSet) {
	mock := &mockSet{}

	ctrl := &controller.Controller{
		Config: controller.Config{
			AwsRegion:       "us-east-0",
			TableName:       "trivy-scan-test",
			S3Region:        "ap-northeast-0",
			S3Bucket:        "example-bucket",
			S3Prefix:        "testing/",
			ScanQueueURL:    "https://sqs.us-east-2.amazonaws.com/123456789012/scan-queue",
			InspectQueueURL: "https://sqs.us-east-2.amazonaws.com/123456789012/inspect-queue",
		},
	}

	ctrl.InjectAdaptors(controller.Adaptors{
		NewDBClient: func(region, tableName string) (interfaces.DBClient, error) {
			var err error
			mock.dbClient, err = db.NewDynamoClientLocal(region, tableName)
			return mock.dbClient, err
		},
	})

	dbClient := ctrl.DB()
	t.Logf("dynamo table name: %s", dbClient.(*db.DynamoClient).TableName())

	t.Cleanup(func() {
		if !t.Failed() {
			if err := ctrl.DB().Close(); err != nil {
				t.Fatal("Can not close DB: ", err)
			}
		}
	})

	return ctrl, mock
}
*/
