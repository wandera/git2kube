## CronJob
* [cronjob.yaml](cronjob.yaml) 
* Deploy git2kube as a Kubernetes CronJob 
* Synchronise with Kubernetes ConfigMap or Secret
* Suitable for longer refresh intervals
* Might be harder to monitor
* Updates might have higher latency due to scheduling
* Low resource requirements

## Watcher
* [watcher.yaml](watcher.yaml) 
* Deploy git2kube as a Kubernetes Deployment
* Synchronise with Kubernetes ConfigMap or Secret
* Suitable for short refresh intervals
* Easier to monitor
* Low latency updates
* Low resource requirements

## Sidecar
* [sidecar.yaml](sidecar.yaml) 
* Deploy git2kube as part of different application Pod
* Synchronise with application by using shared volume
* Suitable for short refresh intervals
* Easier to monitor
* Low latency updates
* Bigger resource requirements