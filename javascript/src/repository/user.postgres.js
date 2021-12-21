import { PasswordNotMatchError, UserNotFoundError } from "../domain/exceptions.js";

export default class UserPostgresRepository {
  constructor(pool) {
    this._pool = pool;
  }
  
  async getById(id) {
    const result = await this._pool.query(
      "SELECT * FROM users WHERE id = $1",
      [id]);

    if (result.rows.length === 0) {
      throw new UserNotFoundError(id);
    }

    return this._toUser(result.rows[0]);
  }

  async create(user) {
    const result = await this._pool.query(
      "INSERT INTO users (username, name, age, surname, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *",
      [user.username, user.name, user.age, user.surname, user.password, new Date(), new Date()]);

    return this._toUser(result.rows[0])
  }

  async updatePassword({id, password, newPassword}) {
    const client = await this._pool.connect();
    try {
      await client.query("BEGIN");

      const result = await this._pool.query(
        "SELECT * FROM users WHERE id = $1",
        [id]);

      if (result.rows.length === 0) {
        throw new UserNotFoundError(id);
      }

      if (result.rows[0].password !== password) {
        throw new PasswordNotMatchError(id);
      }

      await this._pool.query(
        "UPDATE users SET password = $1, updated_at = $3 WHERE id = $2",
        [newPassword, id, new Date()]);
      
      await client.query("COMMIT");

      return { id };
    } catch(error) {
      await client.query("ROLLBACK");
      throw error;
    } finally {
      client.release();
    }
  }

  _toUser(row) {
    const user =  {
      ...row,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }
    delete user.created_at;
    delete user.updated_at;
    return user;
  }

}
