from flask import jsonify, request


class UserController:
    def __init__(self, repository):
        self.repository = repository

    def create(self):
        return jsonify(self.repository.create(request.json))

    def get_by_id(self, id):
        return jsonify(self.repository.get_by_id(id))
