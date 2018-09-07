discord2slack
=============

[![Build Status](https://travis-ci.org/nownabe/discord2slack.svg?branch=master)](https://travis-ci.org/nownabe/discord2slack)


# Development

## Preparation

Set up direnv or configure `APP_PATH` in `docker-compose.yaml`.
`APP_PATH` is application directory path from `GOPATH`.
For example `src/github.com/nownabe/discord2slack`.

Create `app.yaml` and configure environment variables.

```bash
cp .envrc.example .envrc
direnv allow
cp app/app.yaml.example app/app.yaml
vi app/app.yaml
```


## Dev Server

```bash
docker-compose up
```


## Deployment

* Configure `GOOGLE_PROJECT` in `.envrc` or `docker-compose.yaml`
* Create service account JSON key in `$APP_PATH/credentials.yml`

```bash
docker-compose exec app tools/deploy
```


# CI

Required environment variables:

* `GOOGLE_PROJECT`
* `DISCORD_TOKEN`
* `DISCORD_GUILD_IDS`
* `SLACK_WEBHOOK_URL`
