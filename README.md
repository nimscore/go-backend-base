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
grpcurl -plaintext -d '{"slug": "user", "email": "user@example.com", "password": "123456"}' 127.0.0.1:8080 proto.AuthorizationService.Register
grpcurl -plaintext -d '{"email": "user@example.com", "password": "123456"}' 127.0.0.1:8080 proto.AuthorizationService.Login
```

# Schema

```
https://www.drawdb.app/editor?shareId=9febfe9a7f9bce4b6f7be0b63061475f
```
