---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_hook"
sidebar_current: "docs-bitbucket-resource-hook"
description: |-
  Provides a Bitbucket Webhook
---

# bitbucket\_hook

Provides a Bitbucket hook resource.

This allows you to manage your webhooks on a repository.

OAuth2 Scopes: `webhook`

## Example Usage

```hcl
resource "bitbucket_hook" "deploy_on_push" {
  owner       = "myteam"
  repository  = "terraform-code"
  url         = "https://mywebhookservice.mycompany.com/deploy-on-push"
  description = "Deploy the code via my webhook"

  events = [
    "repo:push",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `owner` - (Required) The owner of this repository. Can be you or any team you
  have write access to.
* `repository` - (Required) The name of the repository.
* `url` - (Required) Where to POST to.
* `description` - (Required) The name / description to show in the UI.
* `events` - (Required) The events this webhook is subscribed to. Valid values can be found at [Bitbucket Event Payloads Docs](https://support.atlassian.com/bitbucket-cloud/docs/event-payloads/).

## Import

Hooks can be imported using their `owner/repo-name/hook-id` ID, e.g.

```sh
terraform import bitbucket_hook.hook my-account/my-repo/hook-id
```
