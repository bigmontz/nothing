import { Router } from "express";


function createCrudRouteFor(controller) {
  const router = Router();

  router.get('/:id', controller.getById.bind(controller));
  router.post('/', controller.create.bind(controller));

  return router;
}

export { createCrudRouteFor };
