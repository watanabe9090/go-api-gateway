# Go API Gateway (under development)
A simple api gateway written in go to be used while learning devops


## Developing
You must install nodemon (npm package) `npm i -g nodemon` to be able to run with dev command `make dev`


## Features
- **/api/v1/auth/token**, generates a JWT token if the correct credentials are providen.
- **/api/v1/auth/invalidate** will invalidate the JWT token generated in **/api/v1/auth/token**,
- **/api/v1/\*** will be forward to the end-points defined in the properties files (*props.yaml*)
    - If the end-point requires authentication, 
        - `X-Auth-Username`, username, e.g **foobar**
        - `X-Auth-Role`, role, e.g **ADMIN** or **USER**


## ToDo list
1. ⬜️ Implement password hashing with *bcrypt*
1. ⬜️ Implement *Testcontainers* to be able to make integration tests with *PostgreSQL*
1. ⬜️ Do the tests for tokens/repository.go 
1. ⬜️ Implement the refresh token mechanism 
1. ⬜️ Serach for security improvements
1. ⬜️ Implement Prometheus for monitoring
1. ⬜️ Implement a better logging tool 
1. ✅ Write short documentation
1. ⬜️ Improve the forward route
1. ⬜️ Improve documentation
