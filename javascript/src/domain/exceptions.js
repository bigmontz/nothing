export class UserNotFoundError extends Error {
  constructor(userId) {
    super(`User with id ${userId} not found`);
    this.name = 'UserNotFoundError';
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message
    }
  }
}

export class PasswordNotMatchError extends Error {
  constructor(userId) {
    super(`Password didn't match. (userId: ${userId})`);
    this.name = 'PasswordNotMatchError';
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message
    }
  }
}
