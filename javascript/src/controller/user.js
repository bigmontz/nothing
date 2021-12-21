import { UserNotFoundError, PasswordNotMatchError } from "../domain/exceptions.js";
export default class UserController {
  constructor(userRepository) {
    this._userRepository = userRepository;
  }

  getById = _try(async (req, res) => {
    const user = await this._userRepository.getById(req.params.id);
    res.json(user); 
  })

  create = _try(async (req, res) => {
    const request = {
      username: req.body.username,
      name: req.body.name,
      age: req.body.age,
      surname: req.body.surname,
      password: req.body.password,
    }
    const user = await this._userRepository.create(request);
    res.json(user);
  })

  updatePassword = _try(async (req, res) => {
    const request = {
      id: req.params.id,
      password: req.body.password,
      newPassword: req.body.newPassword
    }
    const user = await this._userRepository.updatePassword(request);
    res.json(user);
  })
}

function _try(fn) {
  return async (req, res) => {
    try {
      await fn(req, res);
    } catch(err) {
      console.error(err);
      if (err instanceof UserNotFoundError) {
        res.status(404).json(err);
      } else if (err instanceof PasswordNotMatchError) {
        res.status(400).json(err);
      } else {
        res.status(500).json(err);
      }
    }
  }
}
