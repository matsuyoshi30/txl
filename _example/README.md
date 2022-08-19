# Example

## setup

MySQL with Docker.

```sh
% docker pull mysql
Using default tag: latest
latest: Pulling from library/mysql
ecf56004177e: Pull complete
1091d9f012e1: Pull complete
fabd9779d467: Pull complete
d6278448e2e1: Pull complete
6050832741b8: Pull complete
a0be9d1181fe: Pull complete
cbdaffd3eb50: Pull complete
a74c1ad0f09f: Pull complete
ad2b7db10f4e: Pull complete
e27624d90cad: Pull complete
ce107dc69a7b: Pull complete
Digest: sha256:ce2ae3bd3e9f001435c4671cf073d1d5ae55d138b16927268474fc54ba09ed79
Status: Downloaded newer image for mysql:latest
docker.io/library/mysql:latest

% docker run -it --name example-mysql -e MYSQL_ROOT_PASSWORD=root -d mysql:latest
25a334996a9c7bee0dccb2791cca78060debfc66e2969e7c142b58dee6dd0995

% docker ps
CONTAINER ID   IMAGE          COMMAND                  CREATED         STATUS         PORTS                 NAMES
25a334996a9c   mysql:latest   "docker-entrypoint.s…"   5 seconds ago   Up 4 seconds   3306/tcp, 33060/tcp   example-mysql

% docker exec -it example-mysql mysql -u root -p -h localhost
Enter password:
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 10
Server version: 8.0.30 MySQL Community Server - GPL

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> 
```

Create database and use.

```sh
mysql> create database recordings;

mysql> use recordings;
```

Create table and insert records.

```sql
DROP TABLE IF EXISTS album;
CREATE TABLE album (
  id         INT AUTO_INCREMENT NOT NULL,
  title      VARCHAR(128) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
  price      DECIMAL(5,2) NOT NULL,
  PRIMARY KEY (`id`)
);

INSERT INTO album
  (title, artist, price)
VALUES
  ('Blue Train', 'John Coltrane', 56.99),
  ('Giant Steps', 'John Coltrane', 63.99),
  ('Jeru', 'Gerry Mulligan', 17.99),
  ('Sarah Vaughan', 'Sarah Vaughan', 34.98);
```

## reference

https://go.dev/doc/tutorial/database-access
