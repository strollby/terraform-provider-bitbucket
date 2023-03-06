---
layout: "bitbucket"
page_title: "Bitbucket: bitbucket_group_members"
sidebar_current: "docs-bitbucket-data-group-members"
description: |-
  Provides a data for Bitbucket group members
---

# bitbucket\_group\_members

Provides a way to fetch data of group members.

## Example Usage

```hcl
data "bitbucket_group_members" "example" {
  workspace = "example"
  slug      = "example"
}
```

## Argument Reference

The following arguments are supported:

* `workspace` - (Required) The UUID that bitbucket groups to connect a group to various objects
* `slug` - (Required) The group's slug.

## Attributes Reference

* `members` - A list of group member uuid.
