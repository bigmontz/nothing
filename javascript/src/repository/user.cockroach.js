
export default class UserCockroachdbRepository {
  constructor(pool) {
    this._pool = pool;
  }

  async getById(id) {
    return await withTransaction(this._pool, async client => {
      const result = await client.query(
        "SELECT * FROM users WHERE id = $1",
        [id]);

      return this._toUser(result.rows[0]);
    });
  }

  async create(user) {
    return await withTransaction(this._pool, async client => {
      const result = await client.query(
        "INSERT INTO users (username, name, age, surname, password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *",
        [user.username, user.name, user.age, user.surname, user.password, new Date(), new Date()]);

      return this._toUser(result.rows[0])
    });
  }

  _toUser(row) {
    const user = {
      ...row,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }
    delete user.created_at;
    delete user.updated_at;
    return user;
  }

}

// Making it simpler
function withTransaction(pool, operation) {
  return new Promise(async (resolve, reject) => {
    const client = await pool.connect();
    retryTxn(0, 15, client,
      async (c, cb) => {
        const result = await operation(c);
        cb(null, result);
      }, async (err, res) => {
        await client.release();
        if (err) {
          reject(err);
        } else {
          resolve(res);
        }
      })
  });
}

// Based on https://www.cockroachlabs.com/docs/stable/build-a-nodejs-app-with-cockroachdb.html?filters=local#step-2-get-the-code
async function retryTxn(n, max, client, operation, callback) {
  await client.query("BEGIN;");
  while (true) {
    n++;
    if (n === max) {
      throw new Error("Max retry count reached.");
    }
    try {
      await operation(client, callback);
      await client.query("COMMIT;");
      return;
    } catch (err) {
      if (err.code !== "40001") {
        return callback(err);
      } else {
        console.log("Transaction failed. Retrying transaction.");
        console.log(err.message);
        await client.query("ROLLBACK;", () => {
          console.log("Rolling back transaction.");
        });
        await new Promise((r) => setTimeout(r, 2 ** n * 1000));
      }
    }
  }
}
