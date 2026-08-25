# Uninstall

Remove the `CostManagementServiceConfig` **before** the operator or the
namespace.

The CR carries a finalizer
(`costmanagementserviceconfigs.service.costmanagement.openshift.io/cleanup`)
that deletes cluster-scoped objects the operator created: the ConsoleLink, and
(only if ROS was enabled) the Kruize ClusterRole and ClusterRoleBinding. That
cleanup runs **only while the operator pod is still running**.

The operator is installed in the **same namespace as the CR**. Deleting that
namespace (or the operator Deployment / CSV) first kills the manager before it
can strip the finalizer. The namespace then stays `Terminating` and the
ConsoleLink leaks cluster-wide.

This does **not** delete your PostgreSQL, Kafka, object storage, or Keycloak.
Those stay until you remove them.

## Procedure

Replace the namespace and CR name with yours.

```bash
export NAMESPACE=cost-onprem
export CR_NAME=cost-management-minimal

# 1. Confirm the operator is still running
oc -n "$NAMESPACE" get deploy,pods

# 2. Delete the CR and wait until it is gone (finalizer cleanup)
oc -n "$NAMESPACE" delete cmsc "$CR_NAME" --timeout=180s

# 3. Confirm NotFound before touching the namespace or operator
oc -n "$NAMESPACE" get cmsc "$CR_NAME"

# 4. Then remove the operator and/or the namespace
oc delete ns "$NAMESPACE"
```

Do not run `oc delete ns` until step 3 returns NotFound. If you installed via
OLM, delete the CR first, then the Subscription / CSV — not the other way
around.

## If the namespace is already Terminating

The operator is gone and the finalizer is stuck. Strip it, then delete the
leaked cluster-scoped objects:

```bash
export NAMESPACE=cost-onprem
export CR_NAME=cost-management-minimal

oc -n "$NAMESPACE" patch cmsc "$CR_NAME" --type=merge \
  -p '{"metadata":{"finalizers":[]}}'
oc delete consolelink "${CR_NAME}-cost-management" --ignore-not-found

# Only if ROS / Kruize was enabled:
oc delete clusterrole,clusterrolebinding \
  -l "app.kubernetes.io/instance=${CR_NAME}" --ignore-not-found
```

The namespace should finish terminating within a few seconds.

## Lab reset

`hack/demo-preprod.sh --reset` and `scripts/install-cmsc.sh` cleanup already
delete the CR first (and strip the finalizer if the operator is already gone).
Cluster Bot tear-down in [clusterbot.md](../development/clusterbot.md) follows
the same order.
