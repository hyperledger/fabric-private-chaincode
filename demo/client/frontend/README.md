<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# frontend

## Project setup
```
make
```

### Compiles and hot-reloads for development
```
make run
```
Note: The service will run on port 5000 and expects the backend to run on http://localhost:3000/api'.
To change the latter, edit the `VUE_APP_API_BASE_URL` variable in the
file `.env.development` accordingly. In that file you can also define
the behaviour of the UI in case of server resets: By default a user
stays logged in but you can for a logout on reset by defining
`VUE_APP_LOGOUT_ON_RESET` to `true`.  Alternatively, you can also
define any these variables as environment variable before starting the front-end.

### Launch docker CLI
```
make cli
```

### CLI commends 

#### Lints and fixes files
```
npm run lint
```
