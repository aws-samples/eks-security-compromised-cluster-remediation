# Isolating the attack

## Cordoning the node

Once you suspect that a pod has been compromised, you want to isolate the pod without alerting the attacker. Cordoning the node that the compromised pod is running on will prevent new pods from being scheduled onto that node. For additional information on cordoning, see [Manual Node Administration](https://kubernetes.io/docs/concepts/architecture/nodes/#manual-node-administration).

### How to cordon a node
To cordon a node, execute the following commands: 

```bash
kubectl cordon <NODE_NAME>
```

## Collecting evidence

Before deleting the compromised pod or the node on which it is running, you need to collect evidence of the breach and formulate a theory about how it occurred. You can start by running [kube-forensics](./kube-forensics.md). kube-forensics is a controller that schedules a Kubernetes Job. The Job runs a series of Docker and `kubectl` commmands and uploads the results to an s3 bucket for analysis. In addition to running kube-forensics, you should also consider taking a [snapshot](https://docs.aws.amazon.com/prescriptive-guidance/latest/backup-recovery/ec2-backup.html) of the instance and capturing the instance's volatile memory. For convenience, we've created an SSM document that installs and runs [AVML](https://github.com/microsoft/avml). The dump file is then pushed to an s3 bucket where it can be downloaded and analyzed using [Volatility](https://github.com/volatilityfoundation/volatility). Volatility is a python application that allows you to run various linux commands, e.g. lsof, bash history, ps, etc against the dump to see what was happening at a point in time. You can also try starting an interactive shell with containers you consider suspect. Before moving on the the next step, determine how and why the attack occurred. 

Creating a forensic workstation and analyzing the dump is out of scope for this workshop. If you're interested in these topics, see the Additional Resources section below.

### Creating and executing the SSM Document

We've created an SSM Document that incorporates the functionality of kube-forensics and creates a memory dump from the instance and uploads it to an s3 bucket. Perform the following actions to create the document in SSM: 

1. Download the `forensics/ssm-document.json` file from the Git repository. If you've already cloned this repository, you can skip this step. 
2. Open a terminal and run the following commands:

    ```bash
    cd ~/environment/eks-security-compromised-cluster-remediation/Containment/forensics
    aws ssm create-document \
    --content file://ssm-document.json \
    --name forensics-capture \
    --document-type Command \
    --document-format JSON \
    --target-type /AWS::EC2::Instance
    ```

3. Run the Document against an instance by executing the following command. Set the environment variables (INSTANCE_ID, NAMESPACE, and POD_NAME) to the actual values.
   You can obtain the EC2 instance ID by running `kubectl get node <NODE_NAME> -o json | jq -r .spec.providerID | cut -d / -f 5`.

    ```bash
    # You will need to set these to the target instance ID, namespace, and pod_name first.
    INSTANCE_ID=...
    NAMESPACE=...
    POD_NAME=...

    aws ssm send-command --document-name "forensics-capture" --document-version "1" \
    --targets '[{"Key":"InstanceIds","Values":["'$INSTANCE_ID'"]}]' \
    --parameters '{"Verbose":["-v"],"TimeoutSeconds":["3600"],"Namespace":["'$NAMESPACE'"],"PodName":["'$POD_NAME'"],"DestinationBucket":["'$FORENSICS_S3_BUCKET'"],"ClusterName":["security-workshop"]}' \
    --timeout-seconds 600 \
    --max-concurrency "50" \
    --max-errors "0" \
    --output-s3-bucket-name "$FORENSICS_S3_BUCKET"
    ```

> The **DestinationBucket** and **output-s3-bucket-name** have been set to the environment variable `FORENSICS_S3_BUCKET` for you. 

### Verification
To verify that the automation script worked, look at the contents of the FORENSICS_S3_BUCKET in the S3 console. The bucket name will start with clusterstack-forensicsbucket. If you see a folder called `forensics`, the script worked as intended. Inside that folder you'll find the memory dump from the instance, an export of the containers in the pod and additional metadata.

<!---
Since the script uploads content to an s3 bucket, the instance on which the script is executed needs s3:PutObject permissions to the destination bucket. Before running the script, add the following inline policy to the instance and/or node group: 

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject"
            ],
            "Resource": "<BUCKET_ARN>/*"
        }
    ]
}
```
-->
### kube-forensics (optional)
kube-forensics obviates the need to specify the node on which the pod is running. Rather, you create a CRD with the pod's name and namespace and the controller will schedule the job on the appropriate node. If you're interested in seeing how kube-forensics works, click [here](./kube-forensics.md) to continue.

## Rescheduling unaffected pods onto other nodes

Once you've cordoned the node, you can begin rescheduling all the unaffected pods running on the node onto other nodes in the cluster. You can do this a couple different ways. The first way is to delete the unaffected pods. A cordoned node is marked as unschedulable. Pods that are deleted from condoned nodes will be rescheduled onto other nodes in the cluster. The second way is to apply a toleration to the compromised pod and then apply a taint to the node. This will evict all pods without a toleration for the taint from the node. For this exercise, we will be cordoning the node and deleting the unaffected pods.

You can use the following bash script to delete the unaffected pods from the node. Be sure to replace the placeholder values `<POD_NAME>` and `<NODE_NAME>` before running the script. 

```bash
#!/bin/bash
BAD_POD=<POD_NAME>
pods=$(kubectl get pods --all-namespaces -o jsonpath='{range .items[*]}{.metadata.name}{"="}{.metadata.namespace}{","}{end}' --field-selector spec.nodeName=<NODE_NAME>)

while read -d, -r pair; do
  IFS='=' read -r key val <<<"$pair"
  if [ $BAD_POD != "$key" ]; then
    kubectl delete pod "$key" -n "$val"
  else
    echo skipping bad pod
  fi
done <<<"$pods,"
```

## Isolating the node (optional)

If you suspect that the node has been compromised, you can isolate it from other nodes by changing its security group. Since Kubernetes version 1.14 and platform version eks.3, a cluster security security group is created by default. This security group allows traffic to flow freely between the control plane and managed node groups. To isolate the compromised node create a security group with the following rules: 

### Node Security Group

| Rule Type | Protocol | Port Range | Source            | Destination      |
|-----------|----------|------------|-------------------|------------------|
| Inbound   | TCP      | 443        | Control plane SG  |                  |
| Inbound   | TCP      | 10250      | Control plane SG  |                  |
| Outbound  | TCP      | 443        |                   | Control plane SG | 

In addition to this SG, you will also need to make the following changes to the cluster security group: 

### Control Plane Security Group

| Rule Type | Protocol | Port Range | Source            | Destination      |
|-----------|----------|------------|-------------------|------------------|
| Inbound   | TCP      | 443        | Node SG           |                  |
| Inbound   | TCP      | 10250      |                   | Node SG          |

Applying these changes will allow the compromised node to continue communicating the with EKS control plane but will prevent it from communicating with other nodes in the cluster. We're allowing the node to communicate with the control plane so as not to alert the attacker. 

## Capture the Flag Challenge (optional)
Find the attacker's Bitcoin account number and [decrypt it](https://www.openssl.org/docs/man1.1.1/man1/openssl-rsautl.html) before deleting the compromised pod(s). 

## Additional Resources

+ [ThreatResponse](https://www.threatresponse.cloud/)
+ [Security Incident: Be Prepared - Memory Dumps - Cloudar](https://www.cloudar.be/awsblog/security-incident-be-prepared-memory-dumps/)
+ [Digital Forensic Analysis of Amazon Linux EC2 Instances](https://www.giac.org/paper/gcfa/13310/digital-forensic-analysis-amazon-linux-ec2-instances/123500)
+ [Rekall](https://github.com/google/rekall)
+ [Volatility](https://github.com/volatilityfoundation/volatility)
+ [AWS_IR](https://aws-ir.readthedocs.io/en/latest/)
+ [MargaritaShotgun](https://margaritashotgun.readthedocs.io/en/latest/)
+ [LiME](https://github.com/504ensicsLabs/LiME)
+ [Linux Memory Forensics - Memory Capture and Analysis](https://youtu.be/6Frec5cGzOg)

## Next Steps
Now that you've mitigated the threat from the compromised pod, it's on to the [Remove and Eradicate](https://github.com/aws-samples/eks-security-compromised-cluster-remediation/tree/main/Eradication_Recovery/remove-compromised-pod) stage of the incident response plan.
