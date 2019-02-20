## Generic Bucket Provisioning

Kubernetes natively supports dynamic provisioning for many types of file and
block storage, but lacks support for object bucket provisioning. 
This repo is a temporary placeholder for bucket provisioning CRDs and related
generated client code. The longer term goal is to move this repo to a Kubernetes
repo such as sig-storage or an external storage repo.

### Design Overview
S3 compatible object store buckets are stable and consistent among most, if not
all, object store solutions. Therefore, the time has come where we can
support a bucket provisioning API similar to that used for Persistent Volumes.
We propose two new Custom Resources to abstract an object store bucket and
a claim/request for such a bucket.  It's important to keep in mind, this
proposal initially only defined bucket and bucket claim APIs and related
client code. See the _*Forward Looking*_ section for more info on how we propose
to augment this initial design with a full bucket library so that all bucket
provisioners can easily guarantee the bucket endpoint and credentials _contract_
described below.

An `ObjectBucketClaim` (OBC) is similar in usage to a Persistent Volume Claim
and an `ObjectBucket` (OB) is the Persistent Volume equivalent. 
Bucket binding refers to the actual bucket being created by the underlying object
store provider. An OBC is namespaced and references a storage class which defines
the object store itself. The details of the object store (ceph, minio, cloud,
on-prem) are not apparent to the developer and can change without disturbing the app.
An OB is non-namespaced (global), typically not visible to end users, and will
contain some info pertinent to the provisioned bucket. Like PVs, there is a 1:1
binding of an OBC to a OB.

Even though OBCs and OBs are generic, the actual bucket provisioning is done by
object store specific operators. For example, if the underlying object store is
AWS S3, the developer creates an OBC, referencing a Storage Class which references
the S3 store. The cluster has the S3 provisioner running which is watching for
OBCs that it knows how to handle. Other OBCs are ignored by the S3 provisioner.
Likewise, the same cluster can also have a rook-ceph RGW provisioner running which also watches OBCs. Like the S3 proivisioner, it only handles OBCs that it knows
how to provision and skips the rest.

In order to access the bucket, the endpoint and key-pair access keys need to be
availble to an application pod, and the pod should not run until the bucket has
been provisioned and can be accessed. This is true even if the pod is created
prior to the OBC.
To synchronize pods with buckets, a ConfigMap will contain the bucket endpoint
and a Secret will contain the key-pair credentials. Pods already block for secrets
and config maps to be mounted, so we have a simple, familiar synchronization pattern.

An OBC can be deleted but the underlying bucket is not removed due to concerns
of deleting objects that cannot be easily recovered. However, OBC deletion triggers
cleanup of Kubernetes resources created on behalf of the bucket, including the secret and config map.
Since the physical bucket is not deleted neither is the OB, which
represents this bucket. The OB's status will indicate that the related OBC
has been deleted so that an admin has better visibility into buckets that are
missing their connection information.

### Details

Bucket binding requires three steps before the bucket is accessible to an app pod:
1. the creation of the physical bucket with the correct owner access key-pairs,
1. the creation of user access key-pairs, in the form of a Secret, granting the app
pod full access to this bucket (but not the ability to create new buckets),
1. the creation of a config map which defines the endpoint of this bucket.

The app pod consumes the config map (endpoint) and secret (key-pairs). The app pod
never sees the OBC or the generated OB.
Note: the provisioner is considered the owner of the bucket, not the OBC or app pod.
This is done to prevent an OBC author from using her access key to create buckets
outside of the Kubernetes cluster. The OBC creator has object PUT, GET, and DELETE
access to the bucket, but cannot create buckets with this access key.

### Controllers

Each object store bucket provisioner is expected to write two separate
reconcillation controllers to support this design. 
For example, if the cluster supports AWS S3 and Rook-Ceph RGW objects stores
then there will be two OBC controllers (one for S3 and one for rook-ceph), and two OB controllers.

1. OBCs in all namespaces are watched and an OB is created when the OBC's Storage
Class' provisioner is recognized by this controller. After the OB operator creates
the bucket, the OBC operator creates the user access Secret and endpoint Config Map.
1.  OBs are watched and a new OB triggers creation of the actual bucket. The 
bucket owner is the object store which has full access.

### Quota

S3 bucket size cannot be specified; however, bucket size can be monitored in S3. The number
of buckets can be controlled by a resource quota once
[k8s pr](https://github.com/kubernetes/kubernetes/pull/72384) is merged. Until then, 
Resource Quotas cannot yet be defined for CRDs.

### Forward Looking

The OB/OBC APIs are enough to support bucket provisioning in Rook-Ceph. However, if
we are going to propose this to the k8s community, there has to be some form of life
cycle manager library for OBCs/OBs. Otherwise we have no way of enforcing the 
behavior/contract described in this proposal. The _binding_ concept would have to be
re-implemented by each provisioner and be fully consistent in order to hide object
store details from the app pod. This also applies to the ConfigMap and Secret being
generated in users’ namespaces with the expected properties, and being deleted 
consistently.

It’s clear that if we want this concept to gain broader k8s acceptance, we need to
present an accompanying lifecycle controller library, similar to the
[sig-storage-lib-external-provisioner](https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner).
Not doing so, and expecting each controller to adhere to the documentation instead, 
likely means the APIs won’t see use outside of the few controllers we write ourselves.

This controller library should define a `Provision()` and `Delete()` interface. The 
`Provision()` method should return an *ObjectBucket, and Credential (defined below), 
while the library consistently generates the Secret and ConfigMap. The approach would
leverage [dynamic informers](https://github.com/kubernetes/kubernetes/pull/69308)

By defining the interfaces and controller framework around these APIs, we establish
a contract between consumers of the APIs and the provisioners who implement them. If 
we do not have the time or resources to follow through on developing this library, 
the value of the APIs is limited almost exclusively to Rook-Ceph. Even in that case,
it would be better to define Rook-Ceph specific APIs in the Rook repo rather
than aiming for a generic solution and temporarily hosting in yard-turkey.

#### When

If we decide that this is the right direction, it seems reasonable that we start
work following the delivery of an initial Rook-Ceph bucket provisioner. This means
that the rook-ceph POC will require a significant rewrite later, but could be
structured to fit an expected interface. This project is not one that can be worked
on in parallel with the rook-ceph implementation as it would be a direct dependency 
of it. Due to time considerations and priorities to support rook-ceph, it cannot be
done before the rook-ceph POC as it would push that timeline too far out in our
opinions.

## API Specifications

### OBC Custom Resource Definition

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbucketclaims.store-operator.s3
spec:
  group: store-operator.s3
  names:
    kind: ObjectBucketClaim
    listKind: ObjectBucketClaimList
    plural: objectbucketclaims
    singular: objectbucketclaim
  scope: Namespaced
  version: v1alpha1
  subresources:
    status: {}
```

### OBC Custom Resource (User Defined)

 ```yaml
apiVersion: objectbucket.s3.io/v1alpha1
kind: ObjectBucketClaim
metadata:
  name: MY-BUCKET-1 [1]
  namespace: USER-NAMESPACE [2]
spec:
  storageClassName: AN-OBJECT-STORE-STORAGE-CLASS [3]
  tenant: MY-TENANT [4]
```
1. name of the ObjectBucketClaim. This name becomes part of the bucket and ConfigMap names.
1. namespace of the ObjectBucketClaim. Determines the namespace of the ConfigMap and user 
Secret. Also becomes part of the unique bucket name.
1. storageClassName is used to target the desired Object Store. Used by the operator to get
the Object Store service URL.
1. tenant allows users to define a tenant in an object store in order to namespace their buckets
and access keys.

### OBC Custom Resource (Status Updated)

 ```yaml
apiVersion: store-operator.s3/v1alpha1
kind: ObjectBucketClaim
...
status:
  phase: {"pending", "bound", "lost"}  [1]
  objectBucketRef: objectReference{}  [2]
  configMapRef: objectReference{}  [3]
  secretRef: objectReference{}  [4]
```
1 `phase` 3 possible phases of bucket creation, mutually exclusive:
    - _pending_: the operator is processing the request
    - _bound_: the operator finished processing the request and linked the OBC and OB
    - _lost_: the OB has been deleted, leaving the OBC unclaimed but unavailable
    TODO: describe how user/admin deals with _lost_ OBs/OBCs
1. `objectBucketRef` is an objectReference to the bound ObjectBucket 
1. `configMapRef` is an objectReference to the generated ConfigMap 
1. `secretRef` is an objectReference to the generated Secret

### OB Custom Resource Definition

 ```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbuckets.store-operator.s3
spec:
  group: store-operator.s3
  names:
    kind: ObjectBucket
    listKind: ObjectBucketList
    plural: objectbuckets
    singular: objectbucket
  scope: Namespaced
  version: v1alpha1
  subresources:
    status: {}
```

 ### Generated OB Custom Resource

 ```yaml
apiVersion: store-operator.s3/v1alpha1
kind: ObjectBucket
metadata:
  name: object-bucket-claim-MY-BUCKET-1
  ownerReferences: [1]
  - name: CEPH-CLUSTER
    ...
  labels:
    ceph.rook.io/object: [2]
spec:
  objectBucketSource: [3]
    provider: ceph.rook.io/object
status:
  claimRef: objectreference [4]
  phase: {"pending", "bound", "lost"} [5]
```
1. `ownerReferences` marks the OB as a child of the object store. If the store is deleted, the bucket will be 
 automatically deleted
1. (optional per provisioner) The label here associates all artifacts under the Rook-Ceph object provisioner
1. `objectBucketSource` is a struct containing metadata of the object store provider
1. `claimRef` is an objectReference to the associated OBC
1. `phase` is the current state of the ObjectBucket:
    - _pending_: the operator is processing the request
    - _bound_: the operator finished processing the request and linked the OBC and OB
    - _lost_: the OBC has been deleted, leaving the OB unclaimed

### Generated Secret for User Access (sample for rook-ceph provider)

 ```yaml
apiVersion: v1
kind: Secret
metadata:
  name: object-bucket-claim-MY-BUCKET-1 [1]
  namespace: USER-NAMESPACE [2]
  labels:
    ceph.rook.io/object: [3]
  ownerReferences:
  - name: MY-BUCKET-1 [4]
    ...
data:
  ACCESS_KEY_ID: BASE64_ENCODED-1
  SECRET_ACCESS_KEY: BASE64_ENCODED-2
```
1. `name` is composed from the OBC's `metadata.name`
1. `namespce` is that of a originating OBC
1. (optional per provisioner) The label here associates all artifacts under the Rook-Ceph object provisioner
1. `ownerReference` makes this secret a child of the originating OBC for clean up purposes

### Generated ConfigMap (sample for rook-ceph provider)

 ```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rook-ceph-object-bucket-MY-BUCKET-1 [1]
  namespace: USER-NAMESPACE [2]
  labels:
    ceph.rook.io/object: [3]
  ownerReferences: [4]
  - name: MY-BUCKET-1
    ...
data: 
  S3_BUCKET_HOST: http://MY-STORE-URL [5]
  S3_BUCKET_PORT: 80 [6]
  S3_BUCKET_NAME: MY-BUCKET-1 [7]
  S3_BUCKET_SSL: no [8]
```
1. `name` composed from `rook-ceph-object-bucket-` and ObjectBucketClaim `metadata.name` value concatenated
1. `namespace` determined by the namespace of the ObjectBucketClaim
1. (optional per provisioner) The label here associates all artifacts under the Rook-Ceph object provisioner
1. `ownerReference` sets the ConfigMap as a child of the ObjectBucketClaim. Deletion of the ObjectBucketClaim causes the deletion of the ConfigMap
1. `S3_BUCKET_HOST` host URL
1. `S3_BUCKET_PORT` host port
1. `S3_BUCKET_NAME` bucket name
1. `S3_BUCKET_SSL` boolean representing SSL connection

### StorageClass (sample for rook-ceph-rgw provider)

 ```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: SOME-OBJECT-STORE
  labels: 
    ceph.rook.io/object: [1]
provisioner: rgw-ceph-rook.io [2]
parameters:
  objectStoreService: MY-STORE [3]
  objectStoreServiceNamespace: MY-STORE-NAMESPACE [4]
  region: LOCATION [5]
```
1. Label `ceph.rook.io/object/claims` associates all artifacts under the ObjectBucketClaim operator.  Defined in example StorageClass and set by cluster admin.  
1. `provisioner` the provisioner responsible to handling OBCs referencing this StorageClass
1. `objectStore` used by the operator to derive the object store Service name.
1. `objectStoreNamespace` the namespace of the object store
1. `region` (optional) defines a region of the object store
