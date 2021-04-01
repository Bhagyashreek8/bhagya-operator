Commands to run this project -

`git clone git@github.com:Bhagyashreek8/bhagya-operator.git`

`cd bhagya-operator`

Make required changes in the code.

Then run go formatting `gofmt -l -w .`

Run below make commands to update the deepcopy and kustomize files.
  `make generate`
  `make manifests`

Build the image `make docker-build IMG=<reponame/imagename:tag>`
Push the image to your repo `make docker-push IMG=<reponame/imagename:tag>`

Update the image in CDR controller deployment spec here -
`controllers/helmchart_controller.go`

Deploy the image to and check if the deployment is created for the operator.

`make deploy IMG=<reponame/imagename:tag>`

to test my current code (half baked)

Pull the image `docker pull bhagyak1/bhagya-test:02`
`make deploy IMG=bhagyak1/bhagya-test:02`


Verify the deployment/pods has been created. This created the CRD deployment.

$ `oc get deployment -n bhagya-operator-system`
```.env
NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
bhagya-operator-controller-manager   1/1     1            1           2m10s
```


$ `oc get pods -n bhagya-operator-system`
```.env
NAME                                                 READY   STATUS    RESTARTS   AGE
bhagya-operator-controller-manager-76cffd554-hh576   2/2     Running   0          2m35s

```


Create a custom resource of type CRD created above.
$ `oc apply -f config/samples/cache_v1_helmchart.yaml` 
helmchart.cache.example.com/bhagyatest-sample created


oc get <CRD_type>
ex- `oc get helmchart`

$ `oc get bhagyatest`
```.env
NAME                AGE
bhagyatest-sample   23s
```


$ `oc get helmchart`
```.env
NAME                AGE
aws-ebs-csi-driver  58s
```


To uninstall the resources completely , run

`make undeploy`





