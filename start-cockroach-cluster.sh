docker-compose up -d cockroachdb_cluster_1 cockroachdb_cluster_2 cockroachdb_cluster_3
docker exec -it cockroachdb_cluster_1 ./cockroach init --insecure
