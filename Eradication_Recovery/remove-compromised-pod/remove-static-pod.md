# Remove static pods from worker nodes

## Background
A Static Pod is a normal Kubernetes Pod that has been configured directly on a cluster node by passing the cluster control plane. A static pod's life-cycle is managed by a node's kubelet rather than the Kubernetes API server and Scheduler. The kubelet will restart a static pod it if it fails. Static pods are typically used to bootstrap a cluster control plane. There are additional use-cases for static pods like storage and network plugins; however, best practice is to use a DaemonSet for applications that require a process on each node in the cluster.

When a static pod is created, the kubelet will create a "mirror" pod on the Kubernetes API server for each static pod. The kubelet will suffix the mirror pod's name with the hostname of the node where the static pod is running. Mirror pods make the static pods visible in the cluster. However, you are unable to use `kubectl` and API Server constructs to manage a static pod.

The kubelet monitors a configured path for addition or removal of pod manifest files. When a pod manifest is added to the directory, the kubelet with start the pod and create a mirror pod on the API Server. Removing the manifest file will cause the kubelet to delete the associated pod and mirror pod.

More information about static pods can be found in the [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/static-pod/ "static pod documentation").

## Prerequisites
We will use your AWS Cloud9 workspace to execute this lab. Here is a list of software that we will use in this lab.

1. [jq](https://www.eksworkshop.com/020_prerequisites/k8stools/#install-jq-envsubst-from-gnu-gettext-utilities-and-bash-completion "jq and other tools")
2. [kubectl](https://www.eksworkshop.com/020_prerequisites/k8stools/#install-kubectl "install kubectl")
3. [The AWS CLI](https://www.eksworkshop.com/020_prerequisites/k8stools/#update-awscli "AWS CLI installation")

## Removing a Static Pod
Before we can remove the static pod we must identify the node(s) running a static pod.

1. Using `kubectl` we can find a static pod by looking for pods with names that include a host-name suffix and that we are unable to delete using standard kubectl commands.

    ```bash
    kubectl get pod -o wide --all-namespaces
    ```

    ```
    NAMESPACE           NAME                                 READY   STATUS    RESTARTS   AGE     IP             NODE                           NOMINATED NODE   READINESS GATES
    sock-shop           carts-844bf95d46-6jj4t               1/1     Running   0          19h     10.0.188.63    ip-10-0-191-117.ec2.internal   <none>           <none>
    sock-shop           carts-db-6c6c68b747-45q45            1/1     Running   0          19h     10.0.175.174   ip-10-0-179-247.ec2.internal   <none>           <none>
    sock-shop           catalogue-759cc6b86-lj5ws            1/1     Running   0          19h     10.0.155.171   ip-10-0-179-247.ec2.internal   <none>           <none>
    sock-shop           catalogue-db-96f6f6b4c-mrn7v         1/1     Running   0          19h     10.0.133.74    ip-10-0-191-117.ec2.internal   <none>           <none>
    sock-shop           front-end-75cf5f5cbc-c8klp           1/1     Running   0          19h     10.0.186.193   ip-10-0-179-247.ec2.internal   <none>           <none>
    sock-shop           orders-7664c64d75-lllzk              1/1     Running   0          19h     10.0.155.232   ip-10-0-191-117.ec2.internal   <none>           <none>
    sock-shop           orders-db-659949975f-xm2mg           1/1     Running   0          19h     10.0.159.186   ip-10-0-179-247.ec2.internal   <none>           <none>
    sock-shop           payment-7bcdbf45c9-r96pw             1/1     Running   0          19h     10.0.232.255   ip-10-0-250-148.ec2.internal   <none>           <none>
    sock-shop           queue-master-5f6d6d4796-j9wx7        1/1     Running   0          19h     10.0.164.241   ip-10-0-179-247.ec2.internal   <none>           <none>
    sock-shop           rabbitmq-57dd566589-q4nwg            2/2     Running   0          16m     10.0.169.201   ip-10-0-191-117.ec2.internal   <none>           <none>
    sock-shop           rship-ip-10-0-250-148.ec2.internal   1/1     Running   0          126m    10.0.250.35    ip-10-0-250-148.ec2.internal   <none>           <none>
    sock-shop           session-db-7cf97f8d4f-nr2nn          1/1     Running   0          19h     10.0.150.22    ip-10-0-179-247.ec2.internal   <none>           <none>
    sock-shop           shipping-7f7999ffb7-5xqq6            1/1     Running   0          19h     10.0.183.255   ip-10-0-191-117.ec2.internal   <none>           <none>
    sock-shop           user-68df64db9c-zb8nj                1/1     Running   0          19h     10.0.158.68    ip-10-0-191-117.ec2.internal   <none>           <none>
    sock-shop           user-db-6df7444fc-67hps              1/1     Running   0          19h     10.0.251.227   ip-10-0-250-148.ec2.internal   <none>           <none>
    ```

    Notice the pod named `rship-ip-10-0-250-148.ec2.internal` in the `sock-shop` NAMESPACE. The pod has an EC2 host-name appended to it, in this case `-ip-10-0-250-148.ec2.internal`.

    This command will report pods where a host-name using the Amazon EC2 private DNS naming convention has been appended to the pod name.
    
    ```bash
    kubectl get pod --all-namespaces | grep -E -o '\s+[a-z].*ip-[0-9]{1,3}\-[0-9]{1,3}\-[0-9]{1,3}\-[0-9]{1,3}\.(ec2|.*\.compute)\.internal\b'
    ```

    ```
    rship-ip-10-0-250-148.ec2.internal
    ```

    If you try to delete that pod, you will notice that you cannot. Notice the age has not decreased and the status is running.

    ```bash
    kubectl -n sock-shop delete pod rship-ip-10-0-250-148.ec2.internal

    kubectl -n sock-shop get pods rship-ip-10-0-250-148.ec2.internal -o wide
    ```

    ```
    NAME                                 READY   STATUS    RESTARTS   AGE    IP            NODE                           NOMINATED NODE   READINESS GATES
    rship-ip-10-0-250-148.ec2.internal   1/1     Running   0          129m   10.0.250.35   ip-10-0-250-148.ec2.internal   <none>           <none>
    ```

2. We will remove the static pod manifest from the worker nodes using [AWS Systems Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/what-is-systems-manager.html "AWS SSM Documentation") (SSM). Once the manifest is removed, the kubelet will delete the static pod. We have provided a simple shell script that uses SSM to execute the Linux `rm` command on the worker nodes in the cluster. In your Cloud9 environment create a new file by selecting "New File" from the "File" menu. Copy and paste the following shell script into the new file. Save the file with the name "rm-static-pods.sh".

    ```bash
    #!/bin/bash
    instances=$(kubectl get nodes -o json | grep providerID | awk -F '/' '{print $NF}' | sed "s/\"//g" | tr '\n' ' ')
    echo "Worker Node Instance IDs:"
    echo $instances

    result=$(aws ssm send-command --document-name "AWS-RunShellScript" --parameters 'commands=["ls -l /etc/kubelet.d ; rm /etc/kubelet.d/* ; ls -l /etc/kubelet.d"]' --instance-ids $instances)
    echo "SSM Command Result:"
    echo $result | jq '.'

    cmd_id=$(echo $result | jq -r '.Command.CommandId')
    echo -n "SSM Command ID:  "
    echo $cmd_id

    echo "pausing for 10 seconds while the commands execute."
    sleep 10
    echo "SSM Command Status"
    for inst in $instances ; do
        echo -n "$inst :> "
        aws ssm get-command-invocation --command-id $cmd_id --instance-id $inst | jq '.Status, .StandardOutputContent, .StandardErrorContent'
    done

    sleep 3
    kubectl get pods -o wide -n sock-shop
    ```

    After saving the file, in your Cloud9 terminal you can run the script with this command

    ```bash
    /bin/bash ./rm-static-pods.sh
    ```

3. Verify the static pod has been deleted. The shell script should end by displaying all running pods in the cluster. You can confirm that the static pod was deleted by reviewing the output of the shell script, or use `kubectl` to get the static pod name identified in step #1.

    ```bash
    kubectl -n sock-shop get pod rship-ip-10-0-250-148.ec2.internal
    ```

    ```
    Error from server (NotFound): pods "rship-ip-10-0-250-148.ec2.internal" not found
    ```

## Next Steps
Now that you've remediated the threat, it's time to implement countermeasures to insulate your cluster from future attacks. Click this [link](https://github.com/aws-samples/eks-security-compromised-cluster-remediation/tree/main/Implement_Countermeasures/gatekeeper) when you're ready to continue. 
