## Instructions

- To play with the go `query_client.go`, execute this command to generate a list of resources
```bash
go run query_client.go > result.yml
```

- Build the go tool locally and use it as cobra command
```bash
make all./bin/odo export all

```

- Next, create a new project/namespace and deploy the list of the k8s resource
```bash
oc new-project dummy
oc create -f result.yaml
```

- Check if all the resources have been well created like the `replicationController` and `Pod`
```bash
oc get all,pvc
NAME                         READY     STATUS    RESTARTS   AGE
pod/my-spring-boot-1-8bfds   1/1       Running   0          3m

NAME                                     DESIRED   CURRENT   READY     AGE
replicationcontroller/my-spring-boot-1   1         1         1         3m

NAME                     TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/my-spring-boot   ClusterIP   172.30.207.131   <none>        8080/TCP   3m

NAME                                                REVISION   DESIRED   CURRENT   TRIGGERED BY
deploymentconfig.apps.openshift.io/my-spring-boot   1          1         1         image(copy-supervisord:latest),image(dev-runtime-spring-boot:latest)

NAME                                                     DOCKER REPO                                               TAGS      UPDATED
imagestream.image.openshift.io/copy-supervisord          172.30.1.1:5000/my-spring-boot1/copy-supervisord          latest    3 minutes ago
imagestream.image.openshift.io/dev-runtime-spring-boot   172.30.1.1:5000/my-spring-boot1/dev-runtime-spring-boot   latest    3 minutes ago

NAME                                           STATUS    VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/m2-data-my-spring-boot   Bound     pv0068    100Gi      RWO,ROX,RWX                   3m
```
