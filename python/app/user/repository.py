
from abc import abstractmethod

from abc import abstractmethod
from datetime import datetime
from psycopg.rows import dict_row
from bson import ObjectId
from pymongo import WriteConcern, ReadPreference
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
            with conn.cursor(row_factory=dict_row) as cursor:
                cursor.execute("SELECT * FROM users WHERE id = %s", (id,))
                row = cursor.fetchone()
                if not row:
                    raise UserNotFoundException()
                return self._to_user(row)

    def create(self, user):
        now = datetime.now()
        query = """INSERT INTO users
                    (username, name, surname,age,
                        password, created_at, updated_at)
                    VALUES (%s, %s, %s, %s, %s, %s, %s) RETURNING *"""
        params = (user["username"], user["name"], user["surname"], user["age"],
                  user["password"], now, now)
        with self.pool.connection() as conn:
            with conn.cursor(row_factory=dict_row) as cursor:
                try:
                    cursor.execute(query, params)
                    row = cursor.fetchone()
                    conn.commit()
                except Exception as e:
                    conn.rollback()
                    raise e
                return self._to_user(row)

    def update_password(self, id, password, new_password):
        # see https://www.psycopg.org/docs/usage.html#transactions-control
        query = """UPDATE users SET updated_at = %s, password = %s
                    WHERE id = %s"""
        params = (datetime.now(), new_password, id)
        with self.pool.connection() as conn:
            with conn.cursor(row_factory=dict_row) as cursor:
                try:
                    cursor.execute("SELECT * FROM users WHERE id = %s", (id,))
                    row = cursor.fetchone()
                    if not row:
                        raise UserNotFoundException()
                    if row["password"] != password:
                        raise PasswordNotMatchException()

                    cursor.execute(query, params)
                    conn.commit()
                except Exception as e:
                    conn.rollback()
                    raise e
                return {"id": row["id"]}

    def _to_user(self, row):
        user = {**row, "createdAt": row["created_at"],
                "updatedAt": row["updated_at"]}
        del user["created_at"]
        del user["updated_at"]
        return user


class UserMongodbRepository(UserRepository):
    def __init__(self, client) -> None:
        super().__init__()
        self.client = client
        self.collection = client.get_database("app").users

    def get_by_id(self, id):
        if not isinstance(id, ObjectId):
            id = ObjectId(id)
        user = self.collection.find_one({"_id": id})
        if not user:
            raise UserNotFoundException()
        return self._to_user(user)

    def create(self, user):
        result = self.collection.insert_one(
            {**user, "createdAt": datetime.now(), "updatedAt": datetime.now()})
        return self.get_by_id(result.inserted_id)

    def update_password(self, id, password, new_password):
        # See https://docs.mongodb.com/manual/core/transactions/
        wc_majority = WriteConcern("majority", wtimeout=1000)
        with self.client.start_session() as session:
            return session.with_transaction(
                self._update_password(id, password, new_password),
                write_concern=wc_majority,
                read_preference=ReadPreference.PRIMARY
            )

    def _update_password(self, id, password, new_password):
        def apply(session):
            if not isinstance(id, ObjectId):
                _id = ObjectId(id)
            else:
                _id = id
            collection = session.client.get_database("app").users
            user = collection.find_one({"_id": _id})
            if not user:
                raise UserNotFoundException()
            if user["password"] != password:
                raise PasswordNotMatchException()

            collection.update_one(
                {"_id": _id},
                {"$set": {"password": new_password,
                          "updatedAt": datetime.now()}})
            return {"id": str(id)}

        return apply

    def _to_user(self, user):
        user = {**user,
                "id": str(user["_id"])}

        del user["_id"]
        return user
