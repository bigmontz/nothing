export default class UserController {
  constructor(userRepository) {
    this._userRepository = userRepository;
  }

  async getById(req, res) {
    const user = await this._userRepository.getById(req.params.id);
    res.json(user); 
  }

  async create(req, res) {
    const request = {
      username: req.body.username,
      name: req.body.name,
      age: req.body.age,
      surname: req.body.surname,
      password: req.body.password,
    }
    const user = await this._userRepository.create(request);
    res.json(user);
  }
}
