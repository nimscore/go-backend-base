# S3 mock

```
sudo pacman -Sy s3cmd
```

```~/.s3cfg
host_bucket = localhost:9090/%(bucket)s
host_domain = localhost:9090
use_https = false
```

```
s3cmd --host=localhost:9090 ls
```

# Minikube

Expose service
```
minikube service -n community community --url
```

# GRPCurl

```
grpcurl -plaintext 127.0.0.1:8080 list
grpcurl -plaintext -d '{"login": "user", "password": "123456"}' 127.0.0.1:8080 iam.IAMService.Login
```
