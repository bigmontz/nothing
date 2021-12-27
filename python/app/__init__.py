from flask import Flask
from os import environ
from .encoder import AppJSONEncoder
from .user import create_blueprint as create_user_blueprint
from .user.repository import UserNeo4jRepository, UserPostgresRepository, UserMongodbRepository, UserCockroachdbRepository


def create_app():
    app = Flask(__name__)
    app.json_encoder = AppJSONEncoder
    repositories = configure_repositories(app)
    app.register_blueprint(create_user_blueprint(
        repositories["user"]), url_prefix="/user")

    return app


def configure_repositories(app):
    factories = {
        "neo4j": configure_neo4j_repository,
        "postgres": configure_postgres_repository,
        "mongodb": configure_mongodb_repository,
        "cockroachdb": configure_cockroachdb_repository
    }
    return factories[environ.get("DB_TYPE", "neo4j")](app)


def configure_neo4j_repository(app):
    from neo4j import GraphDatabase
    driver = GraphDatabase.driver(
        environ.get("NEO4J_URL", "neo4j://localhost"),
        auth=(
            environ.get("NEO4J_USER", "neo4j"),
            environ.get("NEO4J_PASSWORD", "pass")))
    app.teardown_appcontext(lambda _: driver.close())

    return {
        "user": UserNeo4jRepository(driver)
    }


def configure_postgres_repository(app):
    from psycopg_pool import ConnectionPool

    host = environ.get("POSTGRES_URL", "localhost")
    user = environ.get("POSTGRES_USER", "postgres")
    password = environ.get("POSTGRES_PASSWORD", "postgres")
    connection_string = f"postgres://{user}:{password}@{host}/"

    pool = ConnectionPool(connection_string)

    return {
        "user": UserPostgresRepository(pool)
    }


def configure_mongodb_repository(app):
    from pymongo import MongoClient

    host = environ.get("MONGODB_ADDRESS", "localhost")
    user = environ.get("MONGODB_USER", "mongodb")
    password = environ.get("MONGODB_PASSWORD", "mongodb")
    connection_string = f"mongodb://{user}:{password}@{host}/"

    client = MongoClient(connection_string)
    return {
        "user": UserMongodbRepository(client)
    }


def configure_cockroachdb_repository(app):
    from psycopg_pool import ConnectionPool

    host = environ.get("COCKROACH_URL", "localhost")
    user = environ.get("COCKROACH_USER", "admin")
    password = environ.get("COCKROACH_PASSWORD", "cockroach")
    database = environ.get("COCKROACH_DATABASE", "postgres")
    port = environ.get("COCKROACH_PORT", 26257)
    connection_string = f"postgres://{user}:{password}@{host}:{port}/{database}"

    pool = ConnectionPool(connection_string)

    with pool.connection() as conn:
        with conn.cursor() as cursor:
            cursor.execute("""
                CREATE TABLE IF NOT EXISTS users (
                    id SERIAL PRIMARY KEY,
                    username VARCHAR(255) NOT NULL,
                    name VARCHAR(255) NOT NULL,
                    surname VARCHAR(255) NOT NULL,
                    password VARCHAR(255) NOT NULL,
                    age INTEGER NOT NULL,
                    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
                );""")
            conn.commit()

    return {
        "user": UserCockroachdbRepository(pool)
    }
