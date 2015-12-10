# Registry Proxy
Registry-proxy is a simple proxy for docker registry 2.0.
It includes a simple UI to view what containers are stored there and what their tags are.

## Getting Started
Clone the repo and enter into the directory.
Next copy `registry-example.db` to `registry.db`. 
Go to subdirectory `regauth`. Run go install.

> Regauth is a simple command line tool to add users until more features gets added.

Go back to main directory and add a new user to access repo. Run `regauth add [USER]` with [USER] being the username. Enter the password twice and it should work.

Next build the repo by running `make install`
Finally you can deploy

## Roadmap
Change from boltdb to support read/write
Add ability to add/remove users
Add more data than just lists.
Add history of push/pull
Add 404 page.
Add optional ssl
Add kubernetes deployment example
