

export default class UserPostgresRepository {
  constructor(pool) {
    this._pool = pool;
  }
  
  async getById(id) {
    const result = await this._pool.query(
      "SELECT * FROM users WHERE id = $1",
      [id]);

    return this._toUser(result.rows[0]);
  }

  async create(user) {
    const result = await this._pool.query(
      "INSERT INTO users (username, name, age, surname, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *",
      [user.username, user.name, user.age, user.surname, user.password, new Date(), new Date()]);

    return this._toUser(result.rows[0])
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
