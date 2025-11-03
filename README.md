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

grpcurl -plaintext -d '{"name": "user1", "email": "user@example.com", "password": "123456"}' 127.0.0.1:8080 proto.AuthorizationService.Register

grpcurl -plaintext -d '{"email": "user@example.com", "password": "123456"}' 127.0.0.1:8080 proto.AuthorizationService.Login

grpcurl -plaintext -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmF0aW9uIjoxNzYxNzE5MTk3LCJraW5kIjoiYWNjZXNzIiwic2Vzc2lvbl9pZCI6IjIyYWVkOWE1LWQxMzEtNDg0MC05OGZhLTVhMGI0MTAzMWE1MyJ9.PEO0fobI1hzCii8Q1Qr2eHeXZf6FKR50iXJIeZZupEE" 127.0.0.1:8080 proto.AuthorizationService.GetCurrentSession

grpcurl \
    -plaintext \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmF0aW9uIjoxNzYxNzE5NzcxLCJraW5kIjoiYWNjZXNzIiwic2Vzc2lvbl9pZCI6IjUwNjBkYTJhLTVmOTctNDA5Yy1iODQ1LTM0ZjEyZmFkNDUyMyJ9.QtrRKyRQQnJxZSI4HTsOLpfOp_Nv7uKLZGp3VaYSxs0" \
    -d '{"session_id": "5060da2a-5f97-409c-b845-34f12fad4523"}' \
    127.0.0.1:8080 \
    proto.AuthorizationService.RevokeSession
```

```
grpcurl \
    -plaintext \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmF0aW9uIjoxNzYyMTU0MTA3LCJraW5kIjoiYWNjZXNzIiwic2Vzc2lvbl9pZCI6ImVhMThkMTQwLTkzMTAtNDMwZi05NTY0LTIzMDIwYTQ0NWM2OSJ9.2b8AYZlhl74Rc0Vxgysh_P54Xbaez8zO_9Ik8q8osUM" \
    -d '{"name": "name1", "description": "description1", "rules": "rules1"}' \
    127.0.0.1:8080 \
    proto.CommunityService.Create
```

```
grpcurl \
    -plaintext \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHBpcmF0aW9uIjoxNzYyMTU0MTA3LCJraW5kIjoiYWNjZXNzIiwic2Vzc2lvbl9pZCI6ImVhMThkMTQwLTkzMTAtNDMwZi05NTY0LTIzMDIwYTQ0NWM2OSJ9.2b8AYZlhl74Rc0Vxgysh_P54Xbaez8zO_9Ik8q8osUM" \
    -d '{"community_id": "2e96f738-d59e-434c-9708-12a722707cae"}' \
    127.0.0.1:8080 \
    proto.CommunityService.Get
```


# Schema

```
https://www.drawdb.app/editor?shareId=9febfe9a7f9bce4b6f7be0b63061475f
```
