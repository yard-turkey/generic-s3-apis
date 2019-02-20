package controller

import (
	"github.com/yard-turkey/generic-s3-bucket-apis/pkg/controller/objectbucketclaim"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, objectbucketclaim.Add)
}
