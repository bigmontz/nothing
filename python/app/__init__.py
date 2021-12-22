from flask import Flask
from .encoder import AppJSONEncoder
from .user import (
    create_blueprint as create_user_blueprint,
    UserNeo4jRepository
)


def create_app():
    app = Flask(__name__)
    app.json_encoder = AppJSONEncoder
    repositories = configure_repositories(app)
    app.register_blueprint(create_user_blueprint(
        repositories["user"]), url_prefix="/user")

    return app


def configure_repositories(app):
    from neo4j import GraphDatabase
    driver = GraphDatabase.driver(
        "bolt://localhost:7687", auth=("neo4j", "pass"))
    app.teardown_appcontext(lambda _: driver.close())

    return {
        "user": UserNeo4jRepository(driver)
    }
