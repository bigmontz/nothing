from flask import Blueprint
from .routes import register_routes
from .controller import UserController
from .repository import UserRepository, UserNeo4jRepository


def create_blueprint(user_repository: UserRepository):
    blueprint = Blueprint('user', __name__)
    register_routes(blueprint, UserController(user_repository))
    return blueprint
