# eks-security-workshop
## Scenario
As a security engineer at Octank, a global conglomerate, you are responsible for overseeing the security of the compute environment. This includes the computing resources that have been provisioned in the AWS cloud. 

Octank recently launched a new e-commerce site for selling socks. The site, along with all of its dependencies, runs on an EKS cluster in the AWS cloud.  The cluster was initially built and configured without your involvement. Your manager has asked to you verify that the environment meets Octanks stringent security standards. 

Octank has been an appealing target for hackers in the past because of its popularity with consumers and its enormous wealth. You are concerned that the cluster was misconfigured, increasing the probability of a breach. You suspect that the cluster has already been compromised. Your job now is to: 

1. Review the security posture of the cluster
2. Find and isolate the breach
3. Formulate a theory about how or why it occurred
4. Collect evidence for a forensic investigation
5. Eliminate the threat to the environment
6. Implement security controls that decrease or eliminate a recurrence of the issue

## The environment
For this workshop, you will be given access to an AWS account with an EKS cluster. You will access that cluster through a Cloud9 workspace. A slightly modified version of the e-commerce application, [Sock Shop](https://microservices-demo.github.io/), has been deployed to the cluster and the EKS control plane logs have been enabled. Your first priority is to find and isolate the attack. 

The Cloud9 workspace has been granted access to the Kubernetes API and is able to start an SSH session with all the worker nodes in the cluster. An S3 bucket for storing forensic data has also been created and is exposed via the $FORENSICS_S3_BUCKET environment variable.

### The cluster
An EKS cluster has been provisioned in the us-west-2 region. The cluster's 3 worker nodes are distributed across 3 private subnets in the cluster VPC and are part of a [managed node group](https://docs.aws.amazon.com/eks/latest/userguide/managed-node-groups.html). The cluster API endpoint, i.e. the Kubernetes API, is accessible from within the VPC and the Internet. Access to the API server is secured using a combination of AWS Identity and Access Management (IAM) and native Kubernetes Role Based Access Control (RBAC). 

The cluster is accessible from a Cloud9 workspace which has already been provisioned for you. You'll find the workspace under "Shared with you" in the Cloud9 console. 

### The application
Sock Shop is a e-commerce application that consists of a multitude of microservices. The application's front-end is exposed as a LoadBalanced service and is accessible from the Internet. The services communicate with each other as depicted in this diagram:

![](./images/app-architecture.png)

## Instructions
The bulk of this workshop is unguided in that you will decide how to respond to the unfolding security incident at Octank. The workshop’s proctors will be available to periodically give you clues when needed or you can look at the workshop’s [GitHub](https://github.com/aws-samples/eks-security-compromised-cluster-remediation) repository for additional guidance. Once you’ve isolated and/or eliminated the threat from the cluster, you can follow the directions for implementing a few countermeasures that will enhance the security posture of your EKS cluster. This includes implementing OPA Gatekeeper, Falco and Falco Sidekick, and the Security Policy controller.

You will have approximately 2 hours to complete the workshop. 

### Clone this GitHub repository
Start this workshop by cloning this repository to your Cloud9 workspace. 

```bash
git clone https://github.com/aws-samples/eks-security-compromised-cluster-remediation.git
```

## Resources & hints
Aside from this GitHub repository, feel free to use the [EKS Best Practices Guide for Security](https://aws.github.io/aws-eks-best-practices/security/docs/), the official Kubernetes documentation, or other external resources for ideas about how to respond. 

Your first “hint” is to increase your visibility of the cluster and its configuration (or misconfiguration). We recommend using [FairwindsOps/polaris](https://github.com/FairwindsOps/polaris), an open source project from Fairwinds, but you can use another solution if you so choose. See [Detective Controls - EKS Best Practices Guides](https://aws.github.io/aws-eks-best-practices/security/docs/detective/) for a list of potential options.

## Workshop Flow
The workshop is divided into different stages. In the [Identification](./Identification) stage you will implement solutions that will increase your visibility of the cluster, its configuration, and the workloads that are running on it. The idea is to determine whether your cluster has been compromised and how. In the [Containment](./Containment) stage you will isolate the compromise and capture evidence from the environment that can be used in a forensics investigation. In the [Eradication and Recovery](./Eradication_Recovery) stage you will remediate the compromised pod and eliminate the threat to your cluster. And finally in the [Implement Countermeasures](./Implement_Countermeasures) stage you will take stock of what has happened and implement a set of security controls, e.g. OPA/Gatekeeper, Falco, and the Security Profiles Operator, to lower the odds of a recurrence. 

## Capture the flag challenge (optional)
The attacker has left a message for you. If you want an additional challenge, find the attacker's bitcoin account number. 

## Conclusion
You should come away from this workshop with a better sense of: 

1. What controls to put in place to mitigate risks to your Kubernetes clusters
2. How to monitor the environment
3. How to respond when there’s a security incident, e.g. how to collect evidence for a forensics investigation
4. The importance of a good incident response plan
