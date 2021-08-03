/**
 * Describes the IAM authorization details for an AWS service.
 */
export interface ServiceAuthorizationReference {
  /**
   * Name of the service as listed in the service authorization
   * reference.
   */
  name: string;

  /**
   * Prefix seen in IAM action statements for this service.
   */
  servicePrefix: string;

  /**
   * URL of the service authorization reference page for this service.
   */
  authReferenceHref: string;

  /**
   * URL of the API reference for this service, if any.
   */
  apiReferenceHref?: string;

  /**
   * List of actions that can be specified for this service in IAM action statements.
   */
  actions: Action[];

  /**
   * Types of resources that can be specified for this service in IAM resource statements.
   */
  resourceTypes: ResourceType[];

  /**
   * Condition keys that can be specified for this service in IAM statements.
   */
  conditionKeys: ConditionKey[];
}

/**
 * A action that can be allowed or denied via IAM policy.
 */
export interface Action {
  /**
   * Action name as it appears in IAM policy statements.
   */
  name: string;

  /**
   * True if this action is not actually associated with an API call.
   */
  permissionOnly: boolean;

  /**
   * URL of the API or user guide reference for this action.
   */
  referenceHref?: string;

  /**
   * Description of the action.
   */
  description: string;

  /**
   * The access level classification for this action.
   *
   * See the [IAM user guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_understand-policy-summary-access-level-summaries.html) for more information.
   */
  accessLevel: 'List' | 'Read' | 'Write' | 'Permissions management' | 'Tagging';

  /**
   * Resource types that can be specified for this action.
   *
   * If empty, you must specify all resources (`"*"`) in the policy when using this action.
   */
  resourceTypes: ActionResourceType[];
}

/**
 * A resource that can be specified on an action.
 */
export interface ActionResourceType {
  /**
   * A resource type that can be used with an action.
   */
  resourceType: string;

  /**
   * True if a resource of this type is required in order to execute the action.
   * That is, if the IAM statement specifies resources, at least one resource of this type is required.
   */
  required: boolean;

  /**
   * Condition keys that can be specified for this resource type.
   *
   * If a statement specifies a condition key not on this list,
   * and its scope includes a resource of this type, the statement will
   * have no effect.
   */
  conditionKeys: string[];

  /**
   * Additional permissions you must have in order to use the action.
   */
  dependentActions: string[];
}

/**
 * A type of resource that can be specified for a service in an IAM policy.
 */
export interface ResourceType {
  /**
   * Name of the resource type.
   */
  name: string;

  /**
   * URL of the API or user guide reference for this action.
   */
  referenceHref?: string;


  /**
   * Pattern for ARNs for this resource type with `${placeholder}` markers.
   */
  arnPattern: string;

  /**
   * List of condition keys that are valid for this resource type.
   */
  conditionKeys: string[];
}

/**
 * A condition that can be specified for an action in an IAM policy.
 */
export interface ConditionKey {
  /**
   * Name of the condition key, which may contain a template (`${param}`) element.
   */
  name: string;

  /**
   * Link to reference information about the condition key.
   */
  referenceHref?: string;

  /**
   * A short description of the condition key.
   */
  description: string;

  /**
   * The type of the condition key.
   *
   * This can be a primitive type such as String or a compound type such as ArrayOfString.
   */
  type: string;
}

declare const serviceAuth: ServiceAuthorizationReference[];
export { serviceAuth };
