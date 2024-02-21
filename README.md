# smed

Simple editor for [AWS SecretsManager](https://docs.aws.amazon.com/secretsmanager/latest/userguide/intro.html).

Create and edit [json](https://docs.aws.amazon.com/secretsmanager/latest/userguide/reference_secret_json_structure.html) Secrets in your editor.

## Install

    go install github.com/rlister/smed@latest

## Usage

To edit Secrets, set environment variable `EDITOR` to your editor, eg:

    export EDITOR=emacsclient

Edit an existing Secret:

    smed /secret/name

This will open your editor with the json value of the secret, and upload the new value on exit.

Create a new Secret:

    smed -c /new/secret/name

View a Secret as formatted json:

    smed -v /secret/name

List all Secrets:

    smed -l

List all Secrets with name or tags matching either `foo` or `bar`:

    smed -l foo bar
