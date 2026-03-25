output "cluster_id" {
  value = digitalocean_kubernetes_cluster.streamr.id
}

output "cluster_endpoint" {
  value = digitalocean_kubernetes_cluster.streamr.endpoint
}

output "registry_endpoint" {
  value = digitalocean_container_registry.streamr.server_url
}

output "postgres_host" {
  value     = digitalocean_database_cluster.postgres.private_host
  sensitive = true
}

output "postgres_port" {
  value = digitalocean_database_cluster.postgres.port
}

output "valkey_host" {
  value     = digitalocean_database_cluster.valkey.private_host
  sensitive = true
}

output "valkey_port" {
  value = digitalocean_database_cluster.valkey.port
}

output "auth_user_password" {
  value     = digitalocean_database_user.auth_user.password
  sensitive = true
}

output "user_user_password" {
  value     = digitalocean_database_user.user_user.password
  sensitive = true
}

output "stream_user_password" {
  value     = digitalocean_database_user.stream_user.password
  sensitive = true
}

output "chat_user_password" {
  value     = digitalocean_database_user.chat_user.password
  sensitive = true
}

output "notif_user_password" {
  value     = digitalocean_database_user.notif_user.password
  sensitive = true
}

output "valkey_password" {
  value     = digitalocean_database_cluster.valkey.password
  sensitive = true
}