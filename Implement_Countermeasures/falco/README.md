# Introduction
Running multiple containers on the same host exposes the platform to different security threats like _container breakouts_, _noisy neighbor and DOS_, among others. As the popularity of containers and Kubernetes continues to grow, it is increasingly important to have a runtime solution in place to mitigate security threats to your environment. Technologies like _gVisor_ and _Kata Container_ try to isolate containers by sandboxing them, however, this approach has a few limitations as well as performance issues. In this workshop we will look at **Falco** which ensures that containers adheres to a pre-defined set of security standards and quickly remediates potential security threats.

## Falco 
Falco is an opensource runtime security tool originaly built by [Sysdig](https://sysdig.com/). It was donated to CNCF in 2018 and promoted to incubator status in 2020
[GitHub Project](https://github.com/falcosecurity). Falco parses Linux system calls from the kernel at runtime, asserts the stream against a powerful rules engine and alerts when a rule is violated.

### Architecture
Falco consumes events from System Calls via drivers and Kubernetes Audit Events. It has three kind of drivers: 
1. Kernel module: Falco uses kernel module by default and loaded on the host kernel. To install it use this [link](https://falco.org/docs/getting-started/installation/#install-driver)
2. eBPF probe: It is an alternative to kernel module and it brings safety as it doesn't alter the host kernel. eBPF is supported on the newest kernel. To install eBPF probe use [link](https://falco.org/docs/getting-started/installation/#install-driver)
3. Userspace instrumentation: it operates on userspace rather than kernel space.

***
![Falco Architecture](images/falco-extended-architecture.png)

Falco loads its defined rules and scans the event based on the rules in the rule engine. When the filter engine detects a suspicious event, Falco sends the event for alerting to one or many channels: 
- File
- Shell
- Standard Output
- A spawned program
- HTTP/S
- TLS/ gRPC

For details, check documentations [link](https://falco.org/docs/alerts/)

### Rules
Falco rules are written in yaml and follows tcpdump syntax. It contains three types of elements: 
- Rules: Conditions under which an alert should be generated.
- Macros: Rule condition snippets that can be re-used inside rules and even other macros
- List: Collections of items that can be included in rules, macros, or other lists

#### Example:

```
- rule: shell_in_container
  desc: notice shell activity within a container
  condition: evt.type = execve and evt.dir=< and container.id != host and proc.name = bash
  output: shell in a container (user=%user.name container_id=%container.id container_name=%container.name shell=%proc.name parent=%proc.pname cmdline=%proc.cmdline)
  priority: WARNING
```

This rule will alert whenever a shell bash `proc.name = bash` is running inside a container `container.id != host` and successfully spawn of a shell in a container `evt.type = execve and evt.dir=<`.

More details in falco documentation [link](https://falco.org/docs/rules/)

Falco allows a user (security team) to write their own rule, leveraging Falco to alert whenever there is a violation [link](https://falco.org/docs/rules/#appending-to-lists-rules-and-macros)

# Deploy Container Runtime Security using Falco for EKS
In this section, we will deploy Falco to secure container runtime on EKS. Falco could be installed on EC2 instance directly using Linux package manager (yum) or via third party tool like helm to deploy on EKS.

For this workshop, we will use:
- Helm to deploy falco helm chart, use this [link](https://helm.sh/docs/intro/install/) to deploy helm on Cloud9 
- The prebuilt eBPF probe published by Falco Community on [link](https://download.falco.org/). 

If eBPF probe was not published for your AMI version, we can either:
 - build it using [falco documentation](https://falco.org/docs/getting-started/source/) 
 - create a Pull Request on [falcosecurity/test-infra](https://github.com/falcosecurity/test-infra) to add the missing eBPF version
 - Use Bottlereocket instead as it has eBPF probe available by default

## Install Falco Helm Chart
From the Cloud9 workspace: 

```
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update
kubectl create namespace falco
helm install falco falcosecurity/falco -n falco -f values.yaml
```
All the necessary customizations are in the values.yaml files.

```yaml
ebpf:
  # Enable eBPF support for Falco
  enabled: true

auditLog:
  #  Activate the K8s Audit Log feature for Falco
  enabled: true

falco:
  # The location of the rules file(s). This can contain one or more paths to
  # separate rules files.
  rulesFile:
    - /etc/falco/falco_rules.yaml
    - /etc/falco/falco_rules.local.yaml
    - /etc/falco/k8s_audit_rules.yaml
    - /etc/falco/rules.d
  # - /etc/falco/rules.optional.d
 
  # Use json format for output events.
  jsonOutput: true
  jsonIncludeOutputProperty: true
  jsonIncludeTagsProperty: true
  logStderr: true
  logSyslog: true
  logLevel: info
  priority: debug
  
  # Push findings to falcosidekick service
  programOutput:
    enabled: true
    keepAlive: false
    program: "curl -d @- falco-falcosidekick:2801/"

# Deploy Falcosideckick 
falcosidekick:
  enabled: true
  replicaCount: 2
  service:
    type: ClusterIP
    port: 2801
  resources: 
    limits:
      cpu: 100m
      memory: 512Mi
    requests:
      cpu: 15m
      memory: 96Mi
# Use Falcosideckick UI to show falco events
  webui:
    enabled: true
    service:
      # type: NodePort
      type: ClusterIP
      port: 2802
    resources: 
      limits:
        cpu: 200m
        memory: 768Mi
      requests:
        cpu: 50m
        memory: 256Mi
```

Falcosidekick helps to buid a threat response system. It integrates different services like AWS lambda, S3, Slack, etc. Check [here](https://github.com/falcosecurity/falcosidekick#outputs) to see the full list. In this workshop, we use falcosidekick UI to display all the findings. 

### Test
For testing, we will create three suspicious events:

- we create pod and run a command shell inside the container
- we create a pod that will mount `/var/run/docker.sock` as hostpath type socket.
- We create a pod that will mount `/etc/` inside the container as hostpath Directory.

Run tests:

```bash
tests/test.sh
```

### Analyse our finding
After running the tests, inspect the events logged to the falcosidekick UI dashboard:

```bash
kubectl port-forward svc/falco-falcosidekick-ui  8080:2802 -n falco
```

Connect to `https://localhost:8080/ui` from the browser. On the dashboard, you should see a summary for all Falco events captured on EKS Cluster: some events were generated from the EKS nodes where suspicious activities were detected. Other events are surfaced in the dashboard under `Launch Sensitive Mount Container` and `Terminal shell in container`.

Note: kubectl port-forward does not return. You can open a browser instance by clicking preview --> preview running application in Cloud9 IDE to browse the Falco Sidekick UI dashboard.

***

![FalcoSidekick Dashboard](images/falcosidekick-ui-dashboard.png)

***

Click on each Event to check the details:

In this example, we see can information about `k8s.pod.name`, `k8s.ns.name`, `proc.cmdline` which identify the pod, namespace, and what it is trying to run in its shell.

![Shell Output](images/demo-shell.png)

***
In these two examples, we see information like `k8s.pod.name`, `k8s.ns.name`, `container.mounts` which identify the pod and sensitive mounts.

![Hostpath Mount Demo Output](images/demo-hostpath-mount.png)

***

![Hostpath Mount Demo Output](images/demo-docker-sock.png)

***

# Next
Falcosidekick is a powerful threat response system with many integrations. We can leaverage it to quickly excute actions against bad applications, e.g. by deleting pods. Check out this blog series from the Falco Community [link](https://falco.org/blog/falcosidekick-response-engine-part-1-kubeless/) for more information. 

---
