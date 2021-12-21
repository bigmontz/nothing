import { ObjectId } from "mongodb";

export default class UserMongodbRepository {

  constructor(client) {
    this._collection = client.db().collection("users");
  }

  async getById(id) {
    const user = await this._collection.findOne({ _id: new ObjectId(id) });
    return this._toUser(user);
  }

  async create(user) {
    const insertionResult = await this._collection.insertOne({ ...user, createdAt: new Date(), updatedAt: new Date() });
    const insertedUserResult = await this._collection.findOne({ _id: insertionResult.insertedId });
    return this._toUser(insertedUserResult);
  }

  _toUser(insertedUserResult) {
    const user = {
      ...insertedUserResult,
      id: insertedUserResult._id
    };
    delete user._id;
    return user;
  }
}
