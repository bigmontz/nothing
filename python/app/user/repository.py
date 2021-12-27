
from abc import abstractmethod

from abc import abstractmethod
from datetime import datetime
from psycopg.rows import dict_row
from bson import ObjectId
from .exception import PasswordNotMatchException, UserNotFoundException


class UserRepository:
    @abstractmethod
    def create(self, user):
        return user

    @abstractmethod
    def get_by_id(self, id):
        return None


class UserNeo4jRepository(UserRepository):
    def __init__(self, driver) -> None:
        super().__init__()
        self.driver = driver

    def get_by_id(self, id):
        with self.driver.session() as session:
            return session.read_transaction(self._get_by_id, id)

    def _get_by_id(self, tx, id):
        record = tx.run("MATCH (user:User) WHERE ID(user) = $id RETURN user", {
                        "id": int(id)}).single()
        if not record:
            raise UserNotFoundException()
        node = record.get("user")
        return self._to_user(node)

    def create(self, user):
        with self.driver.session() as session:
            return session.write_transaction(self._create, user)

    def _create(self, tx, user):
        now = datetime.now()
        record = tx.run("""CREATE (user:User
                            {
                                username: $username, name: $name,
                                surname: $surname, age: $age,
                                password: $password, createdAt: $createdAt,
                                updatedAt: $updatedAt
                            }) RETURN user""",
                        {**user, "createdAt": now, "updatedAt": now}).single()
        node = record.get("user")
        return self._to_user(node)

    def update_password(self, id, password, new_password):
        with self.driver.session() as session:
            return session.write_transaction(self._update_password, id, password, new_password)

    def _update_password(self, tx, id, password, new_password):
        record = tx.run("MATCH (user:User) WHERE ID(user) = $id RETURN user", {
                        "id": int(id)}).single()
        if not record:
            raise UserNotFoundException()

        node = record.get("user")

        if node.get("password") != password:
            raise PasswordNotMatchException()

        tx.run("""MATCH (user:User) WHERE ID(user) = $id
                SET user.password = $new_password,
                    user.updatedAt=$updated_at""",
               {
                   "id": int(id),
                   "new_password": new_password,
                   "updated_at": datetime.now()
               })

        return {"id": node.id}

    def _to_user(self, node):
        return {
            **node,
            "id": node.id,
            "createdAt": node.get("createdAt").to_native(),
            "updatedAt": node.get("updatedAt").to_native()
        }


class UserPostgresRepository(UserRepository):
    def __init__(self, pool) -> None:
        super().__init__()
        self.pool = pool

    def get_by_id(self, id):
        with self.pool.connection() as conn:
            cursor = conn.cursor(row_factory=dict_row)
            cursor.execute("SELECT * FROM users WHERE id = %s", (id,))
            row = cursor.fetchone()
            return self._to_user(row)

    def create(self, user):
        query = """INSERT INTO users
                    (username, name, surname,age,
                        password, created_at, updated_at)
                    VALUES (%s, %s, %s, %s, %s, %s, %s) RETURNING *"""
        params = (user["username"], user["name"], user["surname"], user["age"],
                  user["password"], datetime.now(), datetime.now())
        with self.pool.connection() as conn:
            cursor = conn.cursor(row_factory=dict_row)
            cursor.execute(query, params)
            row = cursor.fetchone()
            return self._to_user(row)

    def _to_user(self, row):
        user = {**row, "createdAt": row["created_at"],
                "updatedAt": row["updated_at"]}
        del user["created_at"]
        del user["updated_at"]
        return user


class UserMongodbRepository(UserRepository):
    def __init__(self, client) -> None:
        super().__init__()
        self.collection = client.get_database("app").users

    def get_by_id(self, id):
        if not isinstance(id, ObjectId):
            id = ObjectId(id)
        user = self.collection.find_one({"_id": id})
        return self._to_user(user)

    def create(self, user):
        result = self.collection.insert_one(
            {**user, "createdAt": datetime.now(), "updatedAt": datetime.now()})
        return self.get_by_id(result.inserted_id)

    def _to_user(self, user):
        user = {**user,
                "id": str(user["_id"])}

        del user["_id"]
        return user
