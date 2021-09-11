# AWS service authorization reference

This is a JSON-formatted scrape of the [AWS Service Authorization Reference](https://docs.aws.amazon.com/service-authorization/latest/reference/reference.html), along with a Golang program to update it. The package is updated weekly.

## NPM package

If you're using the NPM package, you can use the service reference like this:

```typescript
import { serviceAuth } from '@fluggo/aws-service-auth-reference';

for(const service of serviceAuth) {
  console.log(service.name);
}
```

## Reference

The JSON file contains an array of service reference objects like this:

```javascript
{
  // Name of the service as listed in the service authorization reference.
  "name": "AWS Security Token Service",

  // Prefix seen in IAM action statements for this service.
  "servicePrefix": "sts",

  // URL of the service authorization reference page for this service.
  "authReferenceHref": "https://docs.aws.amazon.com/service-authorization/latest/reference/list_awssecuritytokenservice.html",

  // URL of the API reference for this service, if any.
  "apiReferenceHref": "https://docs.aws.amazon.com/STS/latest/APIReference/",

  // List of actions that can be specified for this service in IAM action statements.
  "actions": [
    {
      // Action name as it appears in IAM policy statements.
      "name": "AssumeRole",

      // True if this action is not actually associated with an API call.
      "permissionOnly": false,

      // URL of the API or user guide reference for this action.
      "referenceHref": "https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html",

      // Description of the action.
      "description": "Returns a set of temporary security credentials that you can use to access AWS resources that you might not normally have access to",

      // The access level classification for this action.
      // This can be List, Read, Write, Permissions management, or Tagging.
      // See https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_understand-policy-summary-access-level-summaries.html
      "accessLevel": "Write",

      // Resource types that can be specified for this action.
      //
      // If empty, you must specify all resources (`"*"`) in the policy when using this action.
      "resourceTypes": [
        {
          // A type of resource that can be used with this action.
          "resourceType": "role",

          // True if at least one resource of this type is required to execute the action.
          "required": true,

          // Condition keys that can be specified for this resource type.
          //
          // If a statement specifies a condition key not on this list,
          // and its scope includes a resource of this type, the statement will
          // have no effect.
          "conditionKeys": [],

          // Additional permissions you must have in order to use the action.
          "dependentActions": []
        }
      ]
    },
    // ...
  ],

  // Types of resources that can be specified for this service in IAM resource statements.
  //
  // These resources can come from other services; check the ARN to see the service type.
  "resourceTypes": [
    {
      // Name of the resource type.
      "name": "role",

      // URL of the API or user guide reference for this action.
      "referenceHref": "https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html",

      // Pattern for ARNs for this resource type with `${placeholder}` markers.
      "arnPattern": "arn:${Partition}:iam::${Account}:role/${RoleNameWithPath}",

      // List of condition keys that are valid for this resource type.
      "conditionKeys": [
        "aws:ResourceTag/${TagKey}"
      ]
    },
    // ...
  ],

  // Condition keys that can be specified for this service in IAM statements.
  "conditionKeys": [
    {
      // Name of the condition key, which may contain a template (`${param}`) element.
      "name": "sts:SourceIdentity",

      // Link to reference information about the condition key.
      "referenceHref": "https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_iam-condition-keys.html#ck_sourceidentity",

      // A short description of the condition key.
      "description": "Filters actions based on the source identity that is passed in the request",

      // The type of the condition key.
      // This can be a primitive type such as String or a compound type such as ArrayOfString.
      "type": "String"
    },
    // ...
  ]
}
```
