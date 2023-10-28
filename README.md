## Config Hierarchy Demo

This project shows how to combine Config file, K8s Config Map and K8s Deployment Env,
this enable config changes without doing code changes and invoke a CI/CD Pipeline
The hierarchy will be like this
[Config File] -> [K8s Config Map] -> [K8s Deployment Env]

### Normal Config File

Start the program by
```bash
$ go run main.go
```

### K8s Config Map

Start the program by
```bash
$ kubectl apply -f deploy/deployment.yaml -n hello
```

### K8s Deployment Env

Start the program by
```bash
$ kubectl apply -f deploy/deployment.yaml -n hello

$ kubectl set env deployment/go-hello-world -n hello NAME=from_set_k8s_env

$ kubectl rollout restart deployment/go-hello-world -n hello
```

### Testing

```bash
$ curl http://<host>:<port>
```

The result will be different for each method used to run the program, 
because we set different value for the program to run, the config file 
will be parsed first, if the config with the same key available in
K8s CM or K8s Env, the value from config file will be overwritten
