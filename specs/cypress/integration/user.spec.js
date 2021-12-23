const PORT =  Cypress.env('API_PORT')  || 3000
const BASE_URL = `http://localhost:${PORT}/user`
const INVALID_ID = Cypress.env('INVALID_ID') || '-1'

describe('User API', () => {

  context('GET /user/{id}', () => {
    let createdUser = null;

    before(() => {
      cy.request('POST', BASE_URL, user())
        .then(response => {
          createdUser = response.body
        })
    });

    it('should get a valid user with the correct property', () => {
      cy.request({
        method: 'GET',
        url: `${BASE_URL}/${createdUser.id}`,
      })
        .should((response) => {
          console.log(response.body)
          console.log(createdUser)
          expect(response.status).to.eq(200)
          expect(response.body.id).to.eq(createdUser.id)
          expect(response.body.username).to.eq(createdUser.username)
          expect(response.body.name).to.eq(createdUser.name)
          expect(response.body.surname).to.eq(createdUser.surname)
          expect(response.body.age).to.eq(createdUser.age)
          expect(response.body.password).to.eq(createdUser.password)
          expect(response.body.createdAt).to.eq(createdUser.createdAt)
          expect(response.body.updatedAt).to.eq(createdUser.updatedAt)
        });
    });

    it('should return 404 for user not found', () => { 
      cy.request({
        method: 'GET',
        url: `${BASE_URL}/${INVALID_ID}`,
        failOnStatusCode: false
      })
        .should((response) => {
          expect(response.status).to.eq(404)
        });
    });
  });

  context('POST /user', () => {
    let testUser = null;
    before(() => {
      testUser = user();

    });

    it('should create a new user', () => {
      cy.request({
        method: 'POST',
        url: `${BASE_URL}`,
        body: testUser
      })
        .should((response) => {
          expect(response.status).to.eq(200)
          expect(response.body.id).to.not.be.null
          expect(response.body.username).to.eq(testUser.username)
          expect(response.body.name).to.eq(testUser.name)
          expect(response.body.surname).to.eq(testUser.surname)
          expect(response.body.age).to.eq(testUser.age)
          expect(response.body.password).to.eq(testUser.password)
          expect(response.body.createdAt).to.eq(response.body.updatedAt)
          // Checking valid date format
          expect(new Date(response.body.createdAt).getTime()).not.eq(NaN)
        });
    });
  });

  context('PUT /user/{id}/password', () => {
    let createdUser = null;

    before(() => {
      cy.request('POST', BASE_URL, user())
        .then(response => {
          createdUser = response.body
        })
    });

    it('should update the password of the user', () => {
      cy.request({
        method: 'PUT',
        url: `${BASE_URL}/${createdUser.id}/password`,
        body: {
          password: createdUser.password,
          newPassword: 'new_password'
        }
      }).should((response) => {
        expect(response.status).to.eq(200)
        expect(response.body.id).to.eq(createdUser.id)
        cy.request({
          method: 'GET',
          url: `${BASE_URL}/${createdUser.id}`,
        }).then((getUserResponse) => {
          expect(getUserResponse.status).to.eq(200)
          expect(getUserResponse.body.password).to.eq('new_password')
        })
      })
    });

    it('should return 404 for user not found', () => {
      cy.request({
        method: 'PUT',
        url: `${BASE_URL}/${INVALID_ID}/password`,
        body: {
          password: createdUser.password,
          newPassword: 'new_password'
        },
        failOnStatusCode: false
      }).should((response) => {
        expect(response.status).to.eq(404)
      })
    });

    it('should return 400 for invalid password', () => {
      cy.request({
        method: 'PUT',
        url: `${BASE_URL}/${createdUser.id}/password`,
        body: {
          password: 'invalid_password',
          newPassword: 'new_password'
        },
        failOnStatusCode: false
      }).should((response) => {
        expect(response.status).to.eq(400)
      })
    });
  });
});

function user() {
  return {
    "username": "the_user_name",
    "name": "The User",
    "surname": "Name",
    "age": Math.ceil((Math.random() * 100)),
    "password": "the_password"
  }
}
