### Go project structure:
- /cmd -> things we run
- /internal -> things we build
## LRU Cache Design
![cache-diagram.png](cache-diagram.png)
The cache provides O(1) access since it uses a 
map from key K to a DLL node (Entry)
which is a (Key,Value) pair and pointers to previous and next element.
The map allows fast lookup to the nodes, while DLL maintains the order for least and
most recently used elements.
The cache keeps a pointer to the head, which represents the most 
recently used element, and a pointer to the tail, 
which represents the least recently used element.

### proto/cache.proto
Defines the gRPC interface for our distributed cache.
Describes the operations which clients can call on nodes.  
The operations for this API:
- Get
- Put
- Remove
- Clear
- Size

Command to generate the protobuf message types and gRPC service definitions.
Run the command from the root of the project.
```
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/cache.proto
```

```
# put a value
grpcurl -plaintext -d '{"key":"name","value":"alice"}' \
  localhost:50051 cache.CacheService/Put

# get it back
grpcurl -plaintext -d '{"key":"name"}' \
  localhost:50051 cache.CacheService/Get

# check size
grpcurl -plaintext -d '{}' \
  localhost:50051 cache.CacheService/Size
```
