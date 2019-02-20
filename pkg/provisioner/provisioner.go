package provisioner

import (
	"github.com/yard-turkey/generic-s3-bucket-apis/pkg/apis"
	"github.com/yard-turkey/generic-s3-bucket-apis/pkg/apis/storeoperator.io/v1alpha1"
	"github.com/yard-turkey/generic-s3-bucket-apis/pkg/controller"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

type Credentials struct {
	AccessKey, SecretKey string
}

type Provisioner interface {
	Provision(*v1alpha1.ObjectBucketClaim) (*v1alpha1.ObjectBucket, *Credentials, error)
	Delete(*v1alpha1.ObjectBucketClaim) error
}

type BucketProvisioner struct {
	client *client.Client
	provisioner Provisioner
	manager manager.Manager
}

func NewBucketProvisioner(client *client.Client, cfg *rest.Config, provisioner Provisioner) (*BucketProvisioner, error) {
	bp := &BucketProvisioner{
		client: client,
		provisioner: provisioner,
	}

	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		return &BucketProvisioner{}, err
	}

	err = apis.AddToScheme(mgr.GetScheme())
	if err != nil {
		return &BucketProvisioner{}, err
	}

	err = controller.AddToManager(mgr)
	if err != nil {
		return &BucketProvisioner{}, err
	}
	bp.manager = mgr
	return bp, nil
}

func (p *BucketProvisioner) Run() {
	if err := p.manager.Start(signals.SetupSignalHandler()); err != nil {
		return
	}
}
