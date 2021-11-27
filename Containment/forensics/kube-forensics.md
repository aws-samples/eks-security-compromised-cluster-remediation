# Installing kube-forensics

> Creates checkpoint snapshots of the state of running pods for later off-line analysis.

kube-forensics allows a cluster administrator to dump the current state of a running pod and all its containers so that security professionals can perform off-line forensic analysis.

In the event of a security breach, members of the Security Team need to examine the state of the Pod and perform a detailed forensics analysis to determine the mode of attack. However, the business would like to terminate the Pod and get back to normal processing as quickly as possible. kube-forensics was developed to allow a cluster administrator to dump the state of a running Pod for offline analysis.

The forensics-controller-manager manages a PodCheckpoint custom resource definition (CRD). The PodCheckpoint resource runs a Kubernetes Job on the same node as the target pod and performs the equivalent of the following operations on the indicated pod/containers by mounting the Docker socket:

``` bash
docker inspect
docker diff
docker export
```

In addition, it collects some meta-data about the target pod by running the equivalent of the following: 

```bash
kubectl describe pod
kubectl get pod
```

The output is uploaded to the destination S3 bucket.

## Installation

You must have cluster administrator access to deploy kube-forensics to a running cluster.

1. Insure your `KUBECONFIG` and current context correctly points to the desired cluster.
1. Checkout kube-forensics repository
1. Change directory into the root of the repository
1. Run `make deploy`

Begin by cloning the Git repository for kube-forensics
```bash
git clone https://github.com/keikoproj/kube-forensics.git
```

Change directories to `/kube-forensics` and build the solution. For example:

```bash
cd kube-forensics
make deploy
```
```
/Users/tekenstam/go/bin/controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
kubectl apply -f config/crd/bases
customresourcedefinition.apiextensions.k8s.io/podcheckpoints.forensics.keikoproj.io configured
kustomize build config/default | kubectl apply -f -
namespace/forensics-system unchanged
customresourcedefinition.apiextensions.k8s.io/podcheckpoints.forensics.keikoproj.io configured
role.rbac.authorization.k8s.io/forensics-leader-election-role unchanged
clusterrole.rbac.authorization.k8s.io/forensics-manager-role configured
clusterrole.rbac.authorization.k8s.io/forensics-proxy-role unchanged
rolebinding.rbac.authorization.k8s.io/forensics-leader-election-rolebinding unchanged
clusterrolebinding.rbac.authorization.k8s.io/forensics-manager-rolebinding unchanged
clusterrolebinding.rbac.authorization.k8s.io/forensics-proxy-rolebinding unchanged
service/forensics-controller-manager-metrics-service unchanged
deployment.apps/forensics-controller-manager unchanged
```

## Usage example

Once the kube-forensics controller is installed, a `PodCheckpoint` spec can be submitted for processing.

<!---
## Update the ServiceAccount to use IRSA

Ordinarily, kube-forensics requires worker nodes to have an IAM role that allows the PutObject API call to an S3 bucket. For this exercise, we are going to update the job's ServiceAccount to use IAM Roles for Service Accounts (IRSA) instead. 

1. Create an S3 bucket. Copy the Arn of the bucket.
2. Download and run the `create-role` binary from the workshop repository. Copy the Arn from the output. 
    
    ```bash
    create-role -bucketArn <BUCKET_ARN> -clusterName <CLUSTER_NAME>
    ```

3.  Annotate the forensics-worker ServiceAccount role with the role Arn. 
    
    ```bash
    kubectl annotate sa forensics-worker -n forensics-system eks.amazonaws.com/role-arn=<ROLE_ARN>
    ```

    Apply this change before creating the custom resource for the job. 

### Sample spec

Save the following `yaml` file to `example.yaml` and modify the `destination`, `pod` and `namespace` to valid values for your cluster.

``` yaml
apiVersion: forensics.keikoproj.io/v1alpha1
kind: PodCheckpoint
metadata:
  name: podcheckpoint-sample
  namespace: forensics-system
spec:
  destination: s3://my-bucket-123456789000-us-west-2
  subpath: forensics
  pod: bad-pod-1234567890-dead1
  namespace: default
```
-->
### Submit & Verify

```bash
kubectl apply -f ./config/samples/forensics_v1alpha1_podcheckpoint.yaml

```
podcheckpoint.forensics.keikoproj.io/podcheckpoint-sample created

```bash
kubectl get -n forensics-system PodCheckpoint
```
```
NAME                   AGE
podcheckpoint-sample   33s
```

Check the state of the PodCheckpoint.

```bash
kubectl describe PodCheckpoint -n forensics-system podcheckpoint-sample
```
```
Name:         podcheckpoint-sample
Namespace:    forensics-system
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"forensics.keikoproj.io/v1alpha1","kind":"PodCheckpoint","metadata":{"annotations":{},"name":"podcheckpoint-sample","namespac...
API Version:  forensics.keikoproj.io/v1alpha1
Kind:         PodCheckpoint
Metadata:
  Creation Timestamp:  2019-08-14T23:19:13Z
  Generation:          2
  Resource Version:    595318
  Self Link:           /apis/forensics.keikoproj.io/v1alpha1/namespaces/forensics-system/podcheckpoints/podcheckpoint-sample
  UID:                 edbe3bd6-bee9-11e9-a5c6-0afa5b77e74c
Spec:
  Destination:  s3://my-bucket-123456789000-us-west-2
  Namespace:    default
  Pod:          bad-pod-1234567890-dead1
  Subpath:      forensics
Status:
  Completion Time:  2019-08-14T23:19:13Z
  Conditions:
    Last Probe Time:       2019-08-14T23:19:13Z
    Last Transition Time:  2019-08-14T23:19:13Z
    Message:               The specified Pod 'bad-pod-1234567890-dead1' was not found in the 'default' namespace.
    Reason:                NotFound
    Status:                True
    Type:                  Failed
  Start Time:              2019-08-14T23:19:13Z
Events:                    <none>
```

In the above output you can see the PodCheckpoint failed due to the Pod name not being found in the system.

## Next Steps
When you're finished experimenting with kube-forensics, click this [link](./README.md) to return to the page on isolating the attack. 
