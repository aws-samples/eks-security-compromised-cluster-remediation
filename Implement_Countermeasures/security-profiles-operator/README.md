## Security Profiles Operator
Linux processes run in [Linux User Space](https://en.wikipedia.org/wiki/User_space), and use [Syscalls](https://man7.org/linux/man-pages/man2/syscalls.2.html) to access Kernel resources in [Linux Kernel Space](https://en.wikipedia.org/wiki/Linux_kernel). Following [least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege) best practices, syscalls should be restricted to only those needed by an application.

Secure computing mode, a.k.a. [seccomp](https://en.wikipedia.org/wiki/Seccomp), is a Linux kernel feature that is used to restrict the syscalls available to the application/container.

In this section of the workshop, attendees will use the `Security Profiles Operator` to install seccomp profiles into the `seccomp-test` Kubernetes namespace, and apply the syscall-restricting profiles to pods running in the `seccomp-test` namespace. The `Security Profiles Operator` installs the seccomp profiles onto cluster nodes. Pods reference the seccomp profiles using the pod-spec `securityContext` element, seen below.

```yaml
  securityContext:
    seccompProfile:
      type: Localhost
      localhostProfile: operator/seccomp-test/<SECCOMP_PROFILE_NAME>.json
```

Applying restrictive seccomp profiles is a way to enforce least privilege syscalls on pods.

(Note: Using a Kubernetes validating admission controller, like [OPA/Gatekeeper](https://kubernetes.io/blog/2019/08/06/opa-gatekeeper-policy-and-governance-for-kubernetes/), can enforce the configuration of seccomp profiles in the `securityContext` element of pod resources.)

---

[Security Profiles Operator](https://github.com/kubernetes-sigs/security-profiles-operator)

[Installation and Usage Documentation](https://github.com/kubernetes-sigs/security-profiles-operator/blob/master/installation-usage.md)

[Restrict a Container's Syscalls with seccomp](https://kubernetes.io/docs/tutorials/clusters/seccomp/)

---

### Installation
```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.4/cert-manager.yaml
kubectl --namespace cert-manager wait --for condition=ready pod -l app.kubernetes.io/instance=cert-manager
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/security-profiles-operator/master/deploy/operator.yaml
```

---

### Examples - Complainer
In this example, a seccomp profile is used to setup seccomp in "complain mode", logging syscalls made by a pod.

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f profiles/ns.yaml
kubectl apply -f profiles/complainer.yaml
```
Using AWS Systems Manager (SSM) Session Manager, start a session with the EKS node EC2 instance, and verify that the seccomp profile was created at `/var/lib/kubelet/seccomp/operator/{NAMESPACE}/{POLICY_NAME}.json` on each node.

```bash
sudo su
cat /var/lib/kubelet/seccomp/operator/seccomp-test/complainer.json
```

```
{"defaultAction":"SCMP_ACT_LOG"}
```

Apply the `complainer` pod, and determine the node on which the pod was scheduled:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f pods/complainer.yaml
kubectl -n seccomp-test get pod complainer -o wide
```

Verify that the NGINX container in the `complainer` pod is running.

Using AWS Systems Manager (SSM) Session Manager, start a session on the pod node, and verify that syscalls are being logged:

```bash
sudo su
cat /var/log/audit/audit.log | grep nginx
```

Log entries should contain syscalls:

```
...
type=SECCOMP msg=audit(1634920756.032:13211): auid=4294967295 uid=0 gid=101 ses=4294967295 pid=19014 comm="nginx" exe="/usr/sbin/nginx" sig=0 arch=c000003e syscall=8 compat=0 ip=0x7f5648cbf597 code=0x7ffc0000
...
```

#### Clean-up

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl delete -f pods/complainer.yaml
kubectl delete -f profiles/complainer.yaml
kubectl delete -f profiles/ns.yaml
```

---

### Examples - Complain and block high-risk
In this example, we will apply a seccomp profile that logs syscalls, blocks high-risk syscalls, and allows some essential syscalls.

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f profiles/ns.yaml
kubectl apply -f profiles/multiple.yaml
```

Using AWS Systems Manager (SSM) Session Manager, start a session with the EKS node EC2 instance, and verify that the seccomp profile was created at `/var/lib/kubelet/seccomp/operator/{NAMESPACE}/{POLICY_NAME}.json` on each node.

```bash
sudo su
ls -al /var/lib/kubelet/seccomp/operator/seccomp-test/
```

```
total 16
drwxr--r-- 2 65535 65535 150 Oct 22 16:30 .
drwxr--r-- 4 65535 65535  60 Oct 21 19:09 ..
-rw-r--r-- 1 65535 65535  34 Oct 22 16:30 profile-allow-unsafe.json
-rw-r--r-- 1 65535 65535  34 Oct 22 16:30 profile-block-all.json
-rw-r--r-- 1 65535 65535 813 Oct 22 16:30 profile-complain-block-high-risk.json
-rw-r--r-- 1 65535 65535  32 Oct 22 16:30 profile-complain-unsafe.json
```
```bash
cat /var/lib/kubelet/seccomp/operator/seccomp-test/profile-complain-block-high-risk.json
```
```
{"defaultAction":"SCMP_ACT_LOG","architectures":["SCMP_ARCH_X86_64"],"syscalls":[{"names":["exit","exit_group","futex","nanosleep"],"action":"SCMP_ACT_ALLOW"},{"names":["acct","add_key","bpf","clock_adjtime","clock_settime","create_module","delete_module","finit_module","get_kernel_syms","get_mempolicy","init_module","ioperm","iopl","kcmp","kexec_file_load","kexec_load","keyctl","lookup_dcookie","mbind","mount","move_pages","name_to_handle_at","nfsservctl","open_by_handle_at","perf_event_open","personality","pivot_root","process_vm_readv","process_vm_writev","ptrace","query_module","quotactl","reboot","request_key","set_mempolicy","setns","settimeofday","stime","swapoff","swapon","_sysctl","sysfs","umount2","umount","unshare","uselib","userfaultfd","ustat","vm86old","vm86"],"action":"SCMP_ACT_ERRNO"}]}
```

Apply the `block-high-risk` pod, and determine the node on which the pod was scheduled:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f pods/block-high-risk.yaml
kubectl -n seccomp-test get pod block-high-risk -o wide
```

Verify that the NGINX container in the `block-high-risk` pod is running.

Using AWS Systems Manager (SSM) Session Manager, start a session on the pod node, and verify that syscalls are being logged:

```bash
sudo su
cat /var/log/audit/audit.log | grep nginx
```

Log entries should contain syscalls:

```bash
...
type=SECCOMP msg=audit(1634920756.032:13211): auid=4294967295 uid=0 gid=101 ses=4294967295 pid=19014 comm="nginx" exe="/usr/sbin/nginx" sig=0 arch=c000003e syscall=8 compat=0 ip=0x7f5648cbf597 code=0x7ffc0000
...
```

#### Clean-up:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl delete -f pods/block-high-risk.yaml
kubectl delete -f profiles/multiple.yaml
kubectl delete -f profiles/ns.yaml
```

---

### Examples - Block all
In this example, we will apply a seccomp profile that blocks all syscalls. This will prevent the NGINX container from starting.

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f profiles/ns.yaml
kubectl apply -f profiles/multiple.yaml
```

Using AWS Systems Manager (SSM) Session Manager, start a session with the EKS node EC2 instance, and verify that the seccomp profile was created at `/var/lib/kubelet/seccomp/operator/{NAMESPACE}/{POLICY_NAME}.json` on each node.

```bash
sudo su
ls -al /var/lib/kubelet/seccomp/operator/seccomp-test/
```
```
total 16
drwxr--r-- 2 65535 65535 150 Oct 22 16:30 .
drwxr--r-- 4 65535 65535  60 Oct 21 19:09 ..
-rw-r--r-- 1 65535 65535  34 Oct 22 16:30 profile-allow-unsafe.json
-rw-r--r-- 1 65535 65535  34 Oct 22 16:30 profile-block-all.json
-rw-r--r-- 1 65535 65535 813 Oct 22 16:30 profile-complain-block-high-risk.json
-rw-r--r-- 1 65535 65535  32 Oct 22 16:30 profile-complain-unsafe.json

```bash
cat /var/lib/kubelet/seccomp/operator/seccomp-test/profile-block-all.json
```
```
{"defaultAction":"SCMP_ACT_ERRNO"}
```

Apply the `block-all` pod, and determine the node on which the pod was scheduled:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f pods/block-all.yaml
kubectl -n seccomp-test get pod block-all -o wide
```

Verify that the NGINX container in the `block-all` pod is not running.
```bash
kubectl -n seccomp-test describe pod block-all | grep -i error
```

The output should be similar to this:
```bash
...
Error: failed to start container "block-all-container": Error response from daemon: cannot start a stopped process: unknown
...
```

#### Clean-up:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl delete -f pods/block-all.yaml
kubectl delete -f profiles/multiple.yaml
kubectl delete -f profiles/ns.yaml
```

---

### Examples - Confine NGINX - Explicit Deny
In this example, we will apply a seccomp profile that was created using the [confine](https://github.com/shamedgh/confine) tool. This profile defaults to the `SCMP_ACT_ALLOW` action and explicitly denies a list of syscalls using the `SCMP_ACT_ERRNO` action.

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f profiles/ns.yaml
kubectl apply -f profiles/confine-nginx-explicit-deny.yaml
```

Using AWS Systems Manager (SSM) Session Manager, start a session with the EKS node EC2 instance, and verify that the seccomp profile was created at `/var/lib/kubelet/seccomp/operator/{NAMESPACE}/{POLICY_NAME}.json` on each node.

```bash
sudo su
ls -al /var/lib/kubelet/seccomp/operator/seccomp-test/
```
```
total 16
total 4
drwxr--r-- 2 65535 65535   32 Oct 27 19:21 .
drwxr--r-- 4 65535 65535   60 Oct 21 19:09 ..
-rw-r--r-- 1 65535 65535 1686 Oct 27 19:21 confine-nginx-explicit-deny.json
```
```bash
cat /var/lib/kubelet/seccomp/operator/seccomp-test/confine-nginx-explicit-deny.json
```
{"defaultAction":"SCMP_ACT_ALLOW","architectures":["SCMP_ARCH_X86_64","SCMP_ARCH_X86","SCMP_ARCH_X32"],"syscalls":[{"names":["msync","mincore","shmctl","pause","getitimer","getpeername","fork","semget","semop","semctl","msgget","msgsnd","msgrcv","msgctl","flock","truncate","symlink","lchown","ptrace","syslog","setreuid","setregid","getpgid","setfsuid","setfsgid","rt_sigpending","rt_sigqueueinfo","utime","mknod","personality","ustat","sysfs","getpriority","sched_rr_get_interval","munlock","mlockall","munlockall","vhangup","modify_ldt","pivot_root","adjtimex","sync","acct","settimeofday","swapon","swapoff","reboot","iopl","ioperm","init_module","delete_module","quotactl","readahead","tkill","set_thread_area","io_submit","io_cancel","get_thread_area","lookup_dcookie","remap_file_pages","restart_syscall","semtimedop","timer_create","timer_settime","timer_gettime","timer_getoverrun","timer_delete","mbind","set_mempolicy","get_mempolicy","mq_open","mq_unlink","mq_timedsend","mq_timedreceive","mq_notify","mq_getsetattr","kexec_load","add_key","ioprio_set","ioprio_get","inotify_init","migrate_pages","mkdirat","mknodat","fchownat","futimesat","renameat","linkat","symlinkat","fchmodat","pselect6","unshare","get_robust_list","tee","sync_file_range","vmsplice","move_pages","signalfd","eventfd","timerfd_gettime","dup3","preadv","rt_tgsigqueueinfo","perf_event_open","recvmmsg","fanotify_init","fanotify_mark","open_by_handle_at","clock_adjtime","syncfs","sendmmsg","getcpu","process_vm_readv","process_vm_writev","kcmp","finit_module","sched_setattr","sched_getattr","renameat2","seccomp","kexec_file_load","bpf","userfaultfd","membarrier","statx"],"action":"SCMP_ACT_ERRNO"}]}
```

Apply the `nginx-deny` pod, and determine the node on which the pod was scheduled:

:point_right: The *kubectl apply* command needs to be executed from the following folder: *eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator*

```bash
kubectl apply -f pods/nginx-explicit-deny.yaml
kubectl -n seccomp-test get pod nginx-deny -o wide
```

Verify that the NGINX container in the `nginx-deny` pod is running.

### Clean-up:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl delete -f pods/nginx-explicit-deny.yaml
kubectl delete -f profiles/confine-nginx-explicit-deny.yaml
kubectl delete -f profiles/ns.yaml
```

---

### Examples - Confine NGINX - Explicit Allow
This profile defaults to the `SCMP_ACT_ERRNO` action and explicitly allows a list of syscalls using the `SCMP_ACT_ALLOW` action.

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f profiles/ns.yaml
kubectl apply -f profiles/confine-nginx-explicit-allow.yaml
```

Using AWS Systems Manager (SSM) Session Manager, start a session with the EKS node EC2 instance, and verify that the seccomp profile was created at `/var/lib/kubelet/seccomp/operator/{NAMESPACE}/{POLICY_NAME}.json` on each node.

```bash
sudo su
ls -al /var/lib/kubelet/seccomp/operator/seccomp-test/
```
```
total 16
total 4
drwxr--r-- 2 65535 65535   32 Oct 27 19:21 .
drwxr--r-- 4 65535 65535   60 Oct 21 19:09 ..
-rw-r--r-- 1 65535 65535 1686 Oct 27 19:21 confine-nginx-explicit-allow.json
```
```bash
cat /var/lib/kubelet/seccomp/operator/seccomp-test/confine-nginx-explicit-allow.json
```
```
{"defaultAction":"SCMP_ACT_ERRNO","architectures":["SCMP_ARCH_X86_64","SCMP_ARCH_X86","SCMP_ARCH_X32"],"syscalls":[{"names":["accept","accept4","access","afs_syscall","alarm","arch_prctl","bind","brk","capget","capset","chdir","chmod","chown","chroot","clock_getres","clock_gettime","clock_nanosleep","clock_settime","clone","close","connect","copy_file_range","creat","create_module","dup","dup2","epoll_create","epoll_create1","epoll_ctl","epoll_ctl_old","epoll_pwait","epoll_wait","epoll_wait_old","eventfd2","execve","execveat","exit","exit_group","faccessat","fadvise64","fallocate","fchdir","fchmod","fchown","fcntl","fdatasync","fgetxattr","flistxattr","fremovexattr","fsetxattr","fstat","fstatfs","fsync","ftruncate","futex","get_kernel_syms","getcwd","getdents","getdents64","getegid","geteuid","getgid","getgroups","getpgrp","getpid","getpmsg","getppid","getrandom","getresgid","getresuid","getrlimit","getrusage","getsid","getsockname","getsockopt","gettid","gettimeofday","getuid","getxattr","inotify_add_watch","inotify_init1","inotify_rm_watch","io_destroy","io_getevents","io_setup","ioctl","keyctl","kill","lgetxattr","link","listen","listxattr","llistxattr","lremovexattr","lseek","lsetxattr","lstat","madvise","memfd_create","mkdir","mlock","mlock2","mmap","mount","mprotect","mremap","munmap","name_to_handle_at","nanosleep","newfstatat","nfsservctl","open","openat","pipe","pipe2","pkey_alloc","pkey_free","pkey_mprotect","poll","ppoll","prctl","pread64","preadv2","prlimit64","putpmsg","pwrite64","pwritev","pwritev2","query_module","read","readlink","readlinkat","readv","recvfrom","recvmsg","removexattr","rename","request_key","rmdir","rt_sigaction","rt_sigprocmask","rt_sigreturn","rt_sigsuspend","rt_sigtimedwait","sched_get_priority_max","sched_get_priority_min","sched_getaffinity","sched_getparam","sched_getscheduler","sched_setaffinity","sched_setparam","sched_setscheduler","sched_yield","security","select","sendfile","sendmsg","sendto","set_robust_list","set_tid_address","setdomainname","setgid","setgroups","sethostname","setitimer","setns","setpgid","setpriority","setresgid","setresuid","setrlimit","setsid","setsockopt","setuid","setxattr","shmat","shmdt","shmget","shutdown","sigaltstack","signalfd","socket","socketpair","splice","stat","statfs","sysinfo","tgkill","time","timerfd_create","timerfd_settime","times","tuxcall","umask","umount2","uname","unlink","unlinkat","uselib","utimensat","utimes","vfork","vserver","wait4","waitid","write","writev"],"action":"SCMP_ACT_ALLOW"}]}
```

Apply the `nginx-allow` pod, and determine the node on which the pod was scheduled:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl apply -f pods/nginx-explicit-allow.yaml
kubectl -n seccomp-test get pod nginx-allow -o wide
```

Verify that the NGINX container in the `nginx-allow` pod is running.

### Clean-up:

```bash
cd ~/environment/eks-security-compromised-cluster-remediation/Implement_Countermeasures/security-profiles-operator/
kubectl delete -f pods/nginx-explicit-allow.yaml
kubectl delete -f profiles/confine-nginx-explicit-allow.yaml
kubectl delete -f profiles/ns.yaml
```

<sub><sup>Note: Seccomp profiles have been adapted from [Security Profiles Operator examples](https://github.com/kubernetes-sigs/security-profiles-operator/tree/master/examples) and [Kubernetes deccomp documentation](https://kubernetes.io/docs/tutorials/clusters/seccomp/).</sup></sub>

---

[Linux x86 64 bit Syscall Table](https://github.com/torvalds/linux/blob/v4.17/arch/x86/entry/syscalls/syscall_64.tbl)

---

[EKS Best Practices Guides](https://aws.github.io/aws-eks-best-practices/)

### Deinstallation (optional)
:warning: Do not delete the operator until you've run through the examples above.

> Note: Profiles cannot be deleted while still in use.

```bash
kubectl delete seccompprofiles --all --all-namespaces
kubectl delete -f https://raw.githubusercontent.com/kubernetes-sigs/security-profiles-operator/master/deploy/operator.yaml
```

### Next Steps
In this section you've seen how you can use seccomp policies to restrict the syscalls a container can issue to the underlying kernel. The profile operator's primary responsibility in this workshop is to deploy seccomp policies onto each of the worker nodes in the cluster so they can be referenced in a pod's securityContext. When controls like seccomp are systematically applied to your containers, it makes it harder for a would-be attacker to escape from a container or hijack it for nefarious purposes. 

When you're ready, click this [link](../falco) to continue to the section on Falco. 
