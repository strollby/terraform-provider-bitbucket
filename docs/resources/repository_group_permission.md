---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_repository_group_permission"
sidebar_current: "docs-bitbucket-resource-repository-group-permission"
description: |-
  Provides a Bitbucket Repository Group Permission Resource
---

# bitbucket\_hook

Provides a Bitbucket Repository Group Permission Resource.

This allows you set explicit group permission for a repository.

OAuth2 Scopes: `repository:admin`

## Example Usage

```hcl
resource "bitbucket_repository_group_permission" "example" {
  workspace  = "example"
  repo_slug  = bitbucket_repository.example.name
  group_slug = bitbucket_group.example.slug
  permission = "read"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The workspace id.
* `repo_slug` - (Required) The repository slug.
* `group_slug` - (Required) Slug of the requested group.
* `permission` - (Required) Permissions can be one of `read`, `write`, and `admin`.

## Import

Repository Group Permissions can be imported using their `workspace:repo-slug:group-slug` ID, e.g.

```sh
terraform import bitbucket_repository_group_permission.example workspace:repo-slug:group-slug
```
