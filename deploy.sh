#!/bin/sh

heroku container:push web -a tv-gin
heroku container:release web -a tv-gin

