# ── Kubernetes Cluster ────────────────────────────────────────────
resource "digitalocean_kubernetes_cluster" "streamr" {
  name    = var.cluster_name
  region  = var.region
  version = "1.35.1-do.0"

  node_pool {
    name       = "worker-pool"
    size       = var.node_size
    node_count = var.node_count

    labels = {
      environment = var.environment
    }
  }
}

# ── VPC (private networking between services and DBs) ─────────────
resource "digitalocean_vpc" "streamr" {
  name   = "streamr-vpc-4"
  region = var.region
}

# ── Managed PostgreSQL Cluster ────────────────────────────────────
resource "digitalocean_database_cluster" "postgres" {
  name       = "streamr-postgres"
  engine     = "pg"
  version    = "16"
  size       = var.db_size
  region     = var.region
  node_count = 1
  private_network_uuid = digitalocean_vpc.streamr.id
}

# ── Individual Databases ──────────────────────────────────────────
resource "digitalocean_database_db" "auth_db" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "auth_db"
}

resource "digitalocean_database_db" "user_db" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "user_db"
}

resource "digitalocean_database_db" "stream_db" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "stream_db"
}

resource "digitalocean_database_db" "chat_db" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "chat_db"
}

resource "digitalocean_database_db" "notification_db" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "notification_db"
}

# ── Database Users ────────────────────────────────────────────────
resource "digitalocean_database_user" "auth_user" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "auth_user"
}

resource "digitalocean_database_user" "user_user" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "user_user"
}

resource "digitalocean_database_user" "stream_user" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "stream_user"
}

resource "digitalocean_database_user" "chat_user" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "chat_user"
}

resource "digitalocean_database_user" "notif_user" {
  cluster_id = digitalocean_database_cluster.postgres.id
  name       = "notif_user"
}

# ── Managed Redis ─────────────────────────────────────────────────
resource "digitalocean_database_cluster" "valkey" {
  name       = "streamr-valkey"
  engine     = "valkey"
  version    = "8"
  size       = "db-s-1vcpu-1gb"
  region     = var.region
  node_count = 1
  private_network_uuid = digitalocean_vpc.streamr.id
}

# ── Firewall: only allow cluster to access DBs ────────────────────
resource "digitalocean_database_firewall" "postgres_fw" {
  cluster_id = digitalocean_database_cluster.postgres.id
  rule {
    type  = "k8s"
    value = digitalocean_kubernetes_cluster.streamr.id
  }
}

resource "digitalocean_database_firewall" "valkey_fw" {
  cluster_id = digitalocean_database_cluster.valkey.id
  rule {
    type  = "k8s"
    value = digitalocean_kubernetes_cluster.streamr.id
  }
}

# ── Container Registry ────────────────────────────────────────────
resource "digitalocean_container_registry" "streamr" {
  name                   = "streamr-registry"
  subscription_tier_slug = "starter"  # free tier
  region                 = var.region
}

resource "digitalocean_container_registry_docker_credentials" "streamr" {
  registry_name = digitalocean_container_registry.streamr.name
}