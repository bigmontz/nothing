import { types } from "neo4j-driver-lite";


export default class UserNeo4jRepository {
  constructor(driver) {
    this._driver = driver;
  }

  async getById(id) {
    const session = this._driver.session();
    try {
      console.log(`Id   ${id}`)
      return await session.readTransaction(async tx => {
        const result = await tx.run(
          `MATCH (user:User) WHERE ID(user) = $id RETURN user`
          , { id: Number(id) });
        return this._nodeToUser(result.records[0].get('user'));
      });
    } finally {
      await session.close();
    }
  }

  async create(user) {
    const session = this._driver.session();
    try {
      return await session.writeTransaction(async tx => {
        const result = await tx.run(
          `CREATE (user:User { username: $username, name: $name, surname: $surname, age: $age, password: $password, createdAt: $createdAt, updatedAt: $updatedAt }) RETURN user`,
          { ...user, createdAt: types.DateTime.fromStandardDate(new Date()), updatedAt: types.DateTime.fromStandardDate(new Date()) }
        );
        return this._nodeToUser(result.records[0].get('user'));
      });

    } finally {
      await session.close()
    }
  }

  _nodeToUser(node) {
    return {
      ...node.properties,
      id: Number(node.identity),
      age: Number(node.properties.age),
      createdAt: node.properties.createdAt.toString(),
      updatedAt: node.properties.updatedAt.toString()
    }
  }
}
