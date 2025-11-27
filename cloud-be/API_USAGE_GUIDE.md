# How to Call the CreateCluster API

## Overview
The `CreateCluster` API is exposed as a **Connect RPC service** (which supports gRPC, gRPC-Web, and REST protocols).

### API Signature
```go
func (h *handlerImpl) CreateCluster(
	ctx context.Context,
	req *connect.Request[k8sv1.CreateClusterRequest],
) (*connect.Response[k8sv1.CreateClusterResponse], error)
```

## Service Details

- **Service Name**: KubernetesService
- **RPC Method**: CreateCluster
- **Package**: `go.bytebuilders.dev/cloud-be/.gen/k8s/v1`
- **Protocol**: Connect RPC (gRPC, gRPC-Web, REST)
- **Default Port**: 8120
- **Endpoint Path**: `/k8s.v1.KubernetesService/CreateCluster`

## How to Call This API

### Option 1: Using Go Client (Connect Client)

```go
package main

import (
	"context"
	"log"

	"connectrpc.com/connect"
	k8sv1 "go.bytebuilders.dev/cloud-be/.gen/k8s/v1"
	"go.bytebuilders.dev/cloud-be/.gen/k8s/v1/k8sv1connect"
)

func main() {
	// Create a new client
	client := k8sv1connect.NewKubernetesServiceClient(
		http.DefaultClient,
		"http://localhost:8120", // Service endpoint
	)

	// Prepare the request
	req := &k8sv1.CreateClusterRequest{
		Name:    "my-cluster",
		Region:  "us-east-1",
		Version: "1.28.0",
		NodePools: []*k8sv1.NodePool{
			{
				Name:  "default",
				Size:  "medium",
				Count: 3,
			},
		},
		Tags: []string{"production", "backend"},
		MaintenancePolicy: &k8sv1.MaintenancePolicy{
			Day:       "Sunday",
			StartTime: "02:00",
		},
	}

	// Call the API
	resp, err := client.CreateCluster(
		context.Background(),
		connect.NewRequest(req),
	)
	if err != nil {
		log.Fatalf("Failed to create cluster: %v", err)
	}

	// Handle response
	cluster := resp.Msg.KubernetesCluster
	log.Printf("Cluster created: ID=%s, Name=%s, Status=%s", 
		cluster.Id, cluster.Name, cluster.Status)
}
```

### Option 2: Using cURL (REST/gRPC-Web)

```bash
# Using gRPC-Web (JSON format)
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-cluster",
    "region": "us-east-1",
    "version": "1.28.0",
    "node_pools": [
      {
        "name": "default",
        "size": "medium",
        "count": 3
      }
    ],
    "tags": ["production", "backend"],
    "maintenance_policy": {
      "day": "Sunday",
      "start_time": "02:00"
    }
  }' \
  http://localhost:8120/k8s.v1.KubernetesService/CreateCluster
```

### Option 3: Using grpcurl (gRPC)

```bash
# First, ensure proto files are available
grpcurl -d '{
  "name": "my-cluster",
  "region": "us-east-1",
  "version": "1.28.0",
  "node_pools": [
    {
      "name": "default",
      "size": "medium",
      "count": 3
    }
  ]
}' \
  localhost:8120 \
  k8s.v1.KubernetesService/CreateCluster
```

### Option 4: Using TypeScript/JavaScript Client

```typescript
import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { KubernetesService } from "./gen/k8s/v1/service_connect";

const transport = createConnectTransport({
  baseUrl: "http://localhost:8120",
});

const client = createClient(KubernetesService, transport);

const response = await client.createCluster({
  name: "my-cluster",
  region: "us-east-1",
  version: "1.28.0",
  nodePools: [
    {
      name: "default",
      size: "medium",
      count: 3,
    },
  ],
  tags: ["production", "backend"],
  maintenancePolicy: {
    day: "Sunday",
    startTime: "02:00",
  },
});

console.log("Cluster created:", response.kubernetesCluster);
```

## Request Structure

The `CreateClusterRequest` message contains:

```
CreateClusterRequest {
  string name                          // Cluster name (required)
  string region                        // Cloud region (required)
  string version                       // Kubernetes version (required)
  repeated NodePool node_pools         // Worker node configurations
  repeated string tags                 // Cluster tags
  MaintenancePolicy maintenance_policy // Maintenance window
  bool auto_upgrade                    // Enable auto-upgrades
  bool surge_upgrade                   // Enable surge upgrades
  bool ha                              // High availability mode
  string vpc_uuid                      // VPC identifier
  // ... additional optional fields
}

NodePool {
  string name              // Pool name
  string size              // Instance size
  int32 count              // Number of nodes
  bool auto_scale          // Enable auto-scaling
  int32 min_nodes          // Minimum nodes (for auto-scaling)
  int32 max_nodes          // Maximum nodes (for auto-scaling)
  map<string, string> labels // Node labels
  repeated string tags      // Pool tags
}

MaintenancePolicy {
  string day         // Day of week (Sunday, Monday, etc.)
  string start_time  // Start time (HH:MM format)
}
```

## Response Structure

The API returns a `CreateClusterResponse`:

```
CreateClusterResponse {
  Cluster kubernetes_cluster  // The created cluster object
}

Cluster {
  string id                   // Unique cluster ID (UUID)
  string name                 // Cluster name
  string region               // Region
  string version              // Kubernetes version
  string endpoint             // Cluster API endpoint (HTTPS)
  string ipv4                 // Cluster IPv4 address
  string cluster_subnet       // Internal cluster subnet
  string service_subnet       // Kubernetes service subnet
  string status               // Provisioning status
  string created_at           // Creation timestamp
  string updated_at           // Last update timestamp
  // ... other fields matching request
}
```

## Response Status Values

The cluster object will have a `status` field with possible values:
- `"provisioning"` - Cluster is being created
- `"ready"` - Cluster is operational
- `"updating"` - Cluster is being updated
- `"deleting"` - Cluster is being deleted
- `"failed"` - Cluster creation failed

## Internal Implementation Details

The API handler internally:

1. **Generates a UUID** for the cluster ID
2. **Creates a Cluster object** with the provided configuration
3. **Triggers a Cadence Workflow** (`KubeVirtClusterCreateWorkflow`) for async provisioning
4. **Returns immediately** with status "provisioning"
5. **Logs the workflow execution** for monitoring

The actual cluster provisioning happens asynchronously via the Cadence workflow engine.

## Error Handling

The API returns Connect RPC errors with appropriate status codes:

```go
if err != nil {
	// The error is a Connect error
	if connect.CodeOf(err) == connect.CodeInternal {
		// Internal server error during workflow start
	}
	// Handle error appropriately
}
```

Common errors:
- `CodeInternal`: Failed to start the provisioning workflow
- `CodeInvalidArgument`: Invalid request parameters
- `CodeUnauthenticated`: Missing or invalid authentication

## Configuration

The service is configured via:
- **Config file**: `config/development.yaml` or `config/base.yaml`
- **Port**: Default 8120 (configurable via `HTTP.Port`)
- **TLS**: Can be enabled via `TLS` configuration section

Example config:
```yaml
services:
  k8s:
    http:
      port: 8120
      bind_address: "0.0.0.0"
    tls:
      enabled: false
```

## Health Check

The service exposes a health endpoint:
```bash
curl http://localhost:8120/k8s/health
# Response: OK (200)
```

## Summary

To call this API, you need to:
1. **Know the endpoint**: `http://localhost:8120` (or configured address)
2. **Use Connect RPC client** (Go, JavaScript, gRPC-Web, etc.)
3. **Send a CreateClusterRequest** with cluster configuration
4. **Receive a CreateClusterResponse** with the cluster object
5. **Monitor the cluster status** via polling or webhook

The actual provisioning happens asynchronously in the background via Cadence workflows.

