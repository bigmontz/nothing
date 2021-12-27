from flask import jsonify, request
from .exception import PasswordNotMatchException, UserNotFoundException


def with_api_wrapper(fn):
    def wrapper(*args, **kwargs):
        try:
            return jsonify(fn(*args, **kwargs))
        except UserNotFoundException as e:
            return jsonify({"error": str(e)}), 404
        except PasswordNotMatchException as e:
            return jsonify({"error": str(e)}), 400
        except Exception as e:
            return jsonify({"error": str(e)}), 500
    return wrapper


class UserController:
    def __init__(self, repository):
        self.repository = repository

    def create(self):
        return with_api_wrapper(self.repository.create)(request.json)

    def get_by_id(self, id):
        return with_api_wrapper(self.repository.get_by_id)(id)

    def update_password(self, id):
        input = request.json
        update = with_api_wrapper(self.repository.update_password)
        return update(id, input["password"], input["newPassword"])
