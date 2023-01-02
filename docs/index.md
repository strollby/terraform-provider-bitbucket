---
layout: "bitbucket"
page_title: "Provider: Bitbucket"
sidebar_current: "docs-bitbucket-index"
description: |-
  The Bitbucket provider to interact with repositories, projects, etc..
---

# Bitbucket Provider

The Bitbucket provider allows you to manage resources including repositories,
webhooks, and default reviewers.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the Bitbucket Provider
provider "bitbucket" {
  username = "GobBluthe"
  password = "idoillusions" # you can also use app passwords
}

resource "bitbucket_repository" "illusions" {
  owner      = "theleagueofmagicians"
  name       = "illusions"
  scm        = "hg"
  is_private = true
}

resource "bitbucket_project" "project" {
  owner      = "theleagueofmagicians" # must be a team
  name       = "illusions-project"
  key        = "ILLUSIONSPROJ"
  is_private = true
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

* `username` - (Optional) Username to use for authentication via [Basic
  Auth](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#basic-auth).
  You can also set this via the `BITBUCKET_USERNAME` environment variable.
  If configured, requires `password` to be configured as well.

* `password` - (Optional) Password to use for authentication via [Basic
  Auth](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#basic-auth).
  Please note that this has to be an [App
  Password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/)
  that has to be created in the [Account
  Settings](https://bitbucket.org/account/settings/app-passwords/). You can
  also set this via the `BITBUCKET_PASSWORD` environment variable. If
  configured, requires `username` to be configured as well.

* `oauth_client_id` - (Optional) OAuth client ID to use for authentication via
  [Client Credentials
  Grant](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#3--client-credentials-grant--4-4-).
  You can also set this via the `BITBUCKET_OAUTH_CLIENT_ID` environment
  variable. If configured, requires `oauth_client_secret` to be configured as
  well.

* `oauth_client_secret` - (Optional) OAuth client secret to use for authentication via
  [Client Credentials
  Grant](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#3--client-credentials-grant--4-4-).
  You can also set this via the `BITBUCKET_OAUTH_CLIENT_SECRET` environment
  variable. If configured, requires `oauth_client_id` to be configured as well.

* `oauth_token` - (Optional) An OAuth access token used for authentication via
  [OAuth](https://developer.atlassian.com/cloud/bitbucket/rest/intro/#oauth-2-0).
  You can also set this via the `BITBUCKET_OAUTH_TOKEN` environment variable.

## OAuth2 Scopes

To interacte with the Bitbucket API, an [App
Password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) or
[OAuth Client
Credentials](https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/)
are required.

App passwords and OAuth client credentials are limited in scope, each API
requires certain scope to interact with, each resource doc will specify what
are the scopes required to use that resource.

See the [Bitbucket OAuth
Documentation](https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/)
for more information on scopes.
