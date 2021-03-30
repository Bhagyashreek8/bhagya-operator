Commands to run this project -

git clone git@github.com:Bhagyashreek8/bhagya-operator.git

cd bhagya-operator

docker pull bhagyak1/bhagya-test:01

make deploy IMG=bhagyak1/bhagya-test:01


Verify the deployment/pods has been created. This created the CRD deployment.

$ oc get deployment -n bhagya-operator-system
NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
bhagya-operator-controller-manager   1/1     1            1           2m10s

$ oc get pods -n bhagya-operator-system
NAME                                                 READY   STATUS    RESTARTS   AGE
bhagya-operator-controller-manager-76cffd554-hh576   2/2     Running   0          2m35s


Create a custom resource of type CRD created above.
$ oc apply -f config/samples/cache_v1_bhagyatest.yaml 
bhagyatest.cache.example.com/bhagyatest-sample created


oc get <CRD_type>
ex- oc get bhagyatest

$ oc get bhagyatest
NAME                AGE
bhagyatest-sample   23s


To uninstall the resources completely , run

make undeploy





