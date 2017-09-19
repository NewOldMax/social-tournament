# Social tournament test task

## Installation

 + make sure that you have installed `docker` and `docker-compose`
 + clone this repository
 + navigate to project forlder
 + rename `.env.dist` to `.env` and change it if needed
 + change `docker-compose.yml` if needed
 + run `docker-compose build` to build containers
 + run `docker-compose up` to start app
 + your app should be available at `127.0.0.1:8099`


## Usage with standalone container

````yaml
version: "2"
services:

    app:
        image: newoldmax/social_tournament
        depends_on:
            - database
        ports:
            - 0.0.0.0:8099:9000
        env_file: .env

    database:
        image: postgres:9.4
        env_file: .env
        environment:
            - PGPASSWORD=example
        ports:
            - "5432"
        volumes:
            - dbdata:/var/lib/postgresql

volumes:
    dbdata:
        driver: local
````

## Possible issues
As `points` have a `float64` type, some rounding errors may happen in cases 100 / 3 = 33.333333333.
It can be fixed to going under `decimal` type and improve calculations based on customer requirements