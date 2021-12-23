from flask import Flask
from os import environ
from .encoder import AppJSONEncoder
from .user import create_blueprint as create_user_blueprint
from .user.repository import UserNeo4jRepository, UserPostgresRepository
from psycopg_pool import ConnectionPool


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
        "postgres": configure_postgres_repository
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
    host = environ.get("POSTGRES_URL", "localhost")
    user = environ.get("POSTGRES_USER", "postgres")
    password = environ.get("POSTGRES_PASSWORD", "postgres")
    connection_string = f"postgres://{user}:{password}@{host}/"

    pool = ConnectionPool(connection_string)

    return {
        "user": UserPostgresRepository(pool)
    }
