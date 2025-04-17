# Pell Node Docker Startup Guide

# 
## Prerequisites

- Docker must be installed and running on your system
- Docker version 20.10.0 or higher recommended

## Startup Process

### 1. Determine Node Type

Choose the network you want to connect to:

- mainnet (Production network)
- testnet (Testing network)

### 2. Fast Sync Configuration

Decide whether to enable fast sync:

- Fast sync: Quicker initial synchronization
- Full sync: Complete blockchain history

### 3. Launch Node

You can start the node in two ways:

#### Option A: Using Docker
Execute the start-docker.sh script:

```bash
./start-docker.sh [NETWORK] [NODE_NAME] [FAST_SYNC]
```

#### Option B: Standard Startup
For a standard startup without Docker, simply run:

```bash
./start-pellcored.sh
```

This script will automatically initialize and start your Pell node with default configurations.
