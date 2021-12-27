
from abc import abstractmethod

from abc import abstractmethod
from datetime import datetime
from psycopg.rows import dict_row
from bson import ObjectId
from pymongo import WriteConcern, ReadPreference
from .exception import PasswordNotMatchException, UserNotFoundException
import logging
import time
from psycopg.errors import SerializationFailure


class UserRepository:
    @abstractmethod
    def create(self, user):
        return user

    @abstractmethod
    def get_by_id(self, id):
        return None

    @abstractmethod
    def update_password(self, id, password, new_password):
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
            return session.write_transaction(self._update_password, id,
                                             password, new_password)

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


# see https://www.cockroachlabs.com/docs/stable/build-a-python-app-with-cockroachdb.html
def run_transaction(conn, op, max_retries=3):
    """
    Execute the operation *op(conn)* retrying serialization failure.

    If the database returns an error asking to retry the transaction, retry it
    *max_retries* times before giving up (and propagate it).
    """
    # leaving this block the transaction will commit or rollback
    # (if leaving with an exception)
    with conn:
        for retry in range(1, max_retries + 1):
            try:

                # If we reach this point, we were able to commit, so we break
                # from the retry loop.

                # THIS PART WAS CHANGED BY ME
                # IT COMMITS AND RETURN THE RESULT
                result = op(conn)
                conn.commit()
                return result

            except SerializationFailure as e:
                # This is a retry error, so we roll back the current
                # transaction and sleep for a bit before retrying. The
                # sleep time increases for each failed transaction.
                logging.debug("got error: %s", e)
                print("got error: %s" % e)
                conn.rollback()
                logging.debug("EXECUTE SERIALIZATION_FAILURE BRANCH")
                sleep_ms = (2 ** retry) * 0.1 * (random.random() + 0.5)
                logging.debug("Sleeping %s seconds", sleep_ms)
                time.sleep(sleep_ms)

            # THIS PART WAS MODIFIED BY ME
            # psycopg2.Error does not exists
            except Exception as e:
                logging.debug("got error: %s", e)
                logging.debug("EXECUTE NON-SERIALIZATION_FAILURE BRANCH")
                raise e

        raise ValueError(
            f"Transaction did not succeed after {max_retries} retries")


class UserCockroachdbRepository(UserRepository):
    def __init__(self, pool) -> None:
        super().__init__()
        self.pool = pool

    def get_by_id(self, id):
        with self.pool.connection() as conn:
            return run_transaction(conn, self._get_by_id(id))

    def _get_by_id(self, id):
        query = "SELECT * FROM users WHERE id = %s"
        params = (id,)

        def apply(conn):
            with conn.cursor(row_factory=dict_row) as cursor:
                cursor.execute(query, params)
                row = cursor.fetchone()
                if not row:
                    raise UserNotFoundException()
                return self._to_user(row)
        return apply

    def create(self, user):
        with self.pool.connection() as conn:
            return run_transaction(conn, self._create(user))

    def _create(self, user):
        now = datetime.now()
        query = """INSERT INTO users
                    (username, name, surname,age,
                        password, created_at, updated_at)
                    VALUES (%s, %s, %s, %s, %s, %s, %s) RETURNING *"""
        params = (user["username"], user["name"], user["surname"], user["age"],
                  user["password"], now, now)

        def apply(conn):
            with conn.cursor(row_factory=dict_row) as cursor:
                cursor.execute(query, params)
                row = cursor.fetchone()
                print(row["id"], type(row["id"]), repr(row["id"]))
                return self._to_user(row)
        return apply

    def update_password(self, id, password, new_password):
        with self.pool.connection() as conn:
            return run_transaction(conn, self._update_password(
                id, password, new_password))

    def _update_password(self, id, password, new_password):
        query = """UPDATE users SET updated_at = %s, password = %s
                    WHERE id = %s"""
        params = (datetime.now(), new_password, id)

        def apply(conn):
            with conn.cursor(row_factory=dict_row) as cursor:
                cursor.execute("SELECT * FROM users WHERE id = %s", (id,))
                row = cursor.fetchone()
                if not row:
                    raise UserNotFoundException()
                if row["password"] != password:
                    raise PasswordNotMatchException()

                cursor.execute(query, params)
                return {"id": row["id"]}

        return apply

    def _to_user(self, row):
        user = {**row, "createdAt": row["created_at"],
                "updatedAt": row["updated_at"]}
        del user["created_at"]
        del user["updated_at"]
        return user
