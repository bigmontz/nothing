#!/usr/bin/env python3
from app import create_app
from flask.cli import FlaskGroup

cli = FlaskGroup(create_app())

if __name__ == "__main__":
    cli()
