# Comparative Costing & Adaptive Scaling Stages

This analysis compares the infrastructure and operational costs of running customer services using a traditional containerized tunnel stack on **Cloud Run v2 (Linux)** versus our **Sovereign WASM (runwasi/gVisor)** architecture across all capacity-adaptive scaling stages.

---

## 1. Cost & Operational Metrics

| Metric Parameter | Legacy Container (Cloud Run v2) | Sovereign WASM (runwasi/gVisor) | Cost & Performance Impact |
| :--- | :---: | :---: | :--- |
| **Artifact Size** | 150MB – 800MB (Base OS + App) | **5MB – 30MB** (Raw WASM) | **>90% reduction** in Artifact Registry storage fees. |
| **Cold Start Latency** | 1,500ms – 5,000ms | **<10ms** | Eliminates latency spikes, providing immediate response without warming. |
| **Minimum Warm Instances** | Required (1 per region to avoid cold starts) | **None** (Zero-Scale Native) | Saves **$15 – $30/month** per service per region in idle CPU charges. |
| **Memory Footprint** | 256MB – 1024MB | **16MB – 64MB** (WASI Slab) | Up to **16x higher density**, cutting underlying compute costs. |
| **Multi-Region Registry Replication** | Eager push of full image layers to all regions | **Adaptive Lazy Promotion** (Only copy when region activates) | Avoids duplicate cross-region network egress charges (~$0.02/GB). |

---

## 2. Capacity-Adaptive Scaling Stages (WASM Stack)

Our controller dynamically adapts compute resource limits, network buffers, and replication routing as workload demands scale up from a single micro-instance:

| Scaling Stage | Target Hardware Profile | Socket Buffer Size | Congestion Algorithm | Target RPS Capacity | Monthly Compute Cost |
| :--- | :--- | :---: | :---: | :---: | :---: |
| **TierSingleRegion** | Local Laptop or single GCP `e2-micro` (2 vCPU, 1GB RAM) | 256 KB | Cubic | 0 – 500 RPS | **$0.00 – $6.00** |
| **TierHorizontalScaling** | Horizontal cluster of `e2-medium` nodes (2 vCPU, 4GB RAM) | 4 MB | Cubic | 500 – 5,000 RPS | **$25.00 – $100.00** |
| **TierMultiRegionSpreading** | Multi-region edge node nodes (us-east1, europe-west3) | 16 MB | Cubic | 5,000 – 25,000 RPS | **$100.00 – $500.00** |
| **TierMeshFederation** | Global federated mesh slices with BBR congestion routing | 32 MB | **BBR** | 25,000+ RPS | **$500.00+** |

---

## 3. Monthly Cost Simulation (5 Regional Endpoints)

Assuming a customer runs **10 microservices** globally across **5 regions** with low-to-medium bursty traffic:

### A. Legacy Container Stack (Cloud Run v2)
*   **Idle Warming Cost**: To prevent cold starts, keeping 1 instance warm per region:
    $$\text{10 services} \times \text{5 regions} \times \$20/\text{month} = \$1,000/\text{month}$$
*   **Storage Fees**: 500MB container images replicated to 5 registries:
    $$\text{10 images} \times 0.5\text{GB} \times \text{5 regions} \times \$0.10/\text{GB-month} = \$2.50/\text{month}$$
*   **Total Monthly Baseline (Before active traffic)**: **~$1,002.50/month**

### B. Sovereign WASM Stack (Native runwasi)
*   **Idle Warming Cost**: **$0.00** (Native WASM instantiates in <10ms, allowing true scale-to-zero).
*   **Storage Fees**: 15MB WASM artifacts replicated adaptively to active registries:
    $$\text{10 artifacts} \times 0.015\text{GB} \times \text{2 active regions} \times \$0.10/\text{GB-month} = \$0.03/\text{month}$$
*   **Total Monthly Baseline**: **~$0.03/month**
