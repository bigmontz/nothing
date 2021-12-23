import { ObjectId } from "mongodb";
import { UserNotFoundError, PasswordNotMatchError } from "../domain/exceptions.js";

export default class UserMongodbRepository {

  constructor(client) {
    this._client = client;
    this._collection = client.db().collection("users");
  }

  async getById(id) {
    const user = await this._collection.findOne({ _id: new ObjectId(id) });
    if (!user) {
      throw new UserNotFoundError(id);
    }
    return this._toUser(user);
  }

  async create(user) {
    const insertionResult = await this._collection.insertOne({ ...user, createdAt: new Date(), updatedAt: new Date() });
    const insertedUserResult = await this._collection.findOne({ _id: insertionResult.insertedId });
    return this._toUser(insertedUserResult);
  }

  async updatePassword({ id, password, newPassword }) {
    const user = await this._collection.findOne({ _id: new ObjectId(id) });
    if (!user) {
      throw new UserNotFoundError(id);
    }
    if(user.password !== password) {
      throw new PasswordNotMatchError(id);
    }
    await this._collection.updateOne({ _id: new ObjectId(id) }, { $set: { password: newPassword, updatedAt: new Date() } });
    return { id };
    // Transations are supported by mongodb, but only in cluster/replicaset mode. So we have different
    // code for running in cluster or local.
    // The solution they use is similar with ours.
    // const session = this._client.startSession();
    // try {
    //   const transactionOptions = {
    //     readPreference: 'primary',
    //     readConcern: { level: 'local' },
    //     writeConcern: { w: 'majority' }
    //   };
    //   // Transaction function...surprise!! :D 
    //   // See https://docs.mongodb.com/manual/core/transactions-in-applications/
    //   return await session.withTransaction(async () => {
    //     const user = await this._collection.findOne({ _id: new ObjectId(id) }, { session });
    //     if (!user) {
    //       throw new UserNotFoundError(id);
    //     }
    //     if(user.password !== password) {
    //       throw new PasswordNotMatchError(id);
    //     }
    //     await this._collection.updateOne({ _id: new ObjectId(id) }, { $set: { password: newPassword, updatedAt: new Date() } }, { session });
    //   }, transactionOptions);
    // } finally {
    //   await session.endSession();
    // } 
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
