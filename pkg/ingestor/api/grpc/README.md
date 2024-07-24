# gRPC api

## Testing
You can trigger a gRPC call by doing this:

```bash
grpcurl -plaintext -format text -d 'cluster_name: "test", run_id: "id"' 127.0.0.1:9000 grpc.API.Ingest
```

Testing rehydrating of all latest scans:
```bash
grpcurl -plaintext -format text 127.0.0.1:9000 grpc.API.RehydrateLatest
```