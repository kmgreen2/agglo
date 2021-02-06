docker run --name postgres -p5432:5432 -e POSTGRESQL_MASTER_PORT_NUMBER=9902 -e POSTGRESQL_PASSWORD=gorm -e POSTGRESQL_USERNAME=gorm -e POSTGRESQL_DATABASE=gorm -d bitnami/postgresql:latest
