package main

import (
	"errors"
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/yuansl/playground/util"
)

func ExamplehuaweiOBS() {
	var bucket string

	obsClient, err := obs.New(_accessKey, _secretKey, "")
	if err != nil {
		util.Fatal("obs.New error:", err)
	}

	result, err := obsClient.GetBucketAcl(bucket)
	if err != nil {
		util.Fatal("obsClient.GetBucketAcl:", err)
	}
	fmt.Printf("owner of bucket %s: '%+v', delivery: '%s'\n", bucket, result.Owner, result.Delivered)

	for i, grant := range result.Grants {
		fmt.Printf("#%d: grant: %+v\n", i, grant)
	}

	// var granted bool

	// if !granted {
	// 	result, err := obsClient.SetBucketAcl(&obs.SetBucketAclInput{
	// 		Bucket: bucket,
	// 		ACL:    obs.AclBucketOwnerFullControl,
	// 		AccessControlPolicy: obs.AccessControlPolicy{
	// 			Owner:  obs.Owner{ID: "7d3547c81e204b91ad7e0f2b0cffc50a"},
	// 			Grants: []obs.Grant{{Grantee: obs.Grantee{ID: "06dfea8249800f1f0f39c016feea94e0"}, Permission: obs.PermissionFullControl}},
	// 		},
	// 	})
	// 	if err != nil {
	// 		util.Fatal("SetBucketAcl error:", err)
	// 	}
	// 	fmt.Printf("SetBucketAcl: %+v\n", result)

	// 	// obsClient.SetObjectAcl(&obs.SetObjectAclInput{})
	// }

	// 获取桶的日志管理配置:
	output, err := obsClient.GetBucketLoggingConfiguration(_BUCKET_NAME)
	if err != nil {
		var obsError obs.ObsError
		switch {
		case errors.As(err, &obsError):
			util.Fatal("obsClient.GetBucketLoggingConfiguration(bucket='%s'): %v\n", _BUCKET_NAME, err)
		default:
			util.Fatal("obsClient.GetBucketLoggingConfiguration(bucket='%s'): unknown error: %v\n", _BUCKET_NAME, err)
		}
	}

	fmt.Printf("Get bucket(%s)'s BucketLoggingConfiguration successful!\n", _BUCKET_NAME)
	fmt.Printf("RequestId:%s\n", output.RequestId)

	fmt.Printf("TargetBucket:%s, TargetPrefix:%s\n", output.TargetBucket, output.TargetPrefix)

	for index, grant := range output.TargetGrants {
		fmt.Printf("Grant[%d]-Type:%s, ID:%s, URI:%s, Permission:%s\n",
			index, grant.Grantee.Type, grant.Grantee.ID, grant.Grantee.URI, grant.Permission)
	}

	// Output: hello
}
