from flask import Blueprint
from .controller import UserController


def register_routes(blueprint: Blueprint, controller: UserController):
    blueprint.route('', methods=['POST'])(controller.create)
    blueprint.route('/<id>', methods=['GET'])(controller.get_by_id)
